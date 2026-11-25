// Importa os imóveis do parceiro RTA Consultoria Imobiliária (São Paulo).
//
// O site é pequeno e estático: a página inicial lista os imóveis com os dados
// em atributos data-* (tipo, dormitórios, vagas, valor, área, bairro) e cada
// página de imóvel traz a rua, as suítes e a galeria (data-base/data-total).
//
// Uso (com a API local no ar):
//   node importador/importar-rta.mjs --simular
//   node importador/importar-rta.mjs
//
// Variáveis: API_URL (padrão http://localhost:8080), ADMIN_EMAIL, ADMIN_SENHA.

const SITE = "https://rtaimoveis.pages.dev";
const API = process.env.API_URL || "http://localhost:8080";
const MAX_FOTOS = 12;
const PAUSA_MS = 1000;
const SIMULAR = process.argv.includes("--simular");

// O site não informa a cidade; os bairros (Moema, Campo Belo, Jardim
// América, Paraíso…) e o telefone (11) são de São Paulo.
const CIDADE = "São Paulo";
const ESTADO = "SP";

const UA = "BretasImoveis-Importador/1.0 (parceria RTA)";
const dormir = (ms) => new Promise((r) => setTimeout(r, ms));

const decodificar = (s = "") =>
  s
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&nbsp;/g, " ")
    .trim();

async function baixarTexto(url) {
  const r = await fetch(url, { headers: { "User-Agent": UA }, redirect: "follow" });
  if (!r.ok) throw new Error(`${r.status} em ${url}`);
  return r.text();
}

const num = (v) => Math.round(Number(String(v || "").replace(",", ".")) || 0);

// ── Página inicial: um card por imóvel ─────────────────────────────────────
function lerCards(html) {
  const cards = [];
  for (const m of html.matchAll(/<a class="card"([^>]*)>([\s\S]*?)<\/a>/g)) {
    const attrs = Object.fromEntries(
      [...m[1].matchAll(/([\w-]+)="([^"]*)"/g)].map((a) => [a[1], decodificar(a[2])])
    );
    const rua = decodificar(m[2].match(/<span class="end">([^<]*)</)?.[1]);
    cards.push({
      href: attrs.href,
      slug: attrs.href.replace(/^.*\//, "").replace(/\.html$/, ""),
      tipo: attrs["data-tipo"] || "Apartamento",
      quartos: num(attrs["data-dorms"]),
      vagas: num(attrs["data-vagas"]),
      venda: num(attrs["data-venda"]),
      locacao: num(attrs["data-locacao"]),
      area: num(attrs["data-m2"]),
      // O parceiro escreve alguns bairros sem acento.
      bairro: ({ Paraiso: "Paraíso" })[attrs["data-bairro"]] || attrs["data-bairro"] || "",
      rua,
    });
  }
  return cards;
}

// ── Página do imóvel: suítes e galeria ─────────────────────────────────────
function lerDetalhe(html, urlPagina) {
  const suites = num(html.match(/sendo\s+(\d+)\s+su[ií]te/i)?.[1]);
  const g = html.match(/class="grade-fotos"[^>]*data-base="([^"]+)"[^>]*data-total="(\d+)"/);
  const fotos = [];
  if (g) {
    const base = new URL(g[1] + "/", urlPagina);
    const total = Math.min(Number(g[2]), MAX_FOTOS);
    for (let i = 1; i <= total; i++) fotos.push(new URL(`${String(i).padStart(2, "0")}.jpg`, base).href);
  }
  return { suites, fotos };
}

function montarImovel(card, det) {
  // Tipos do parceiro → os três tipos do nosso site.
  const tipo = /casa/i.test(card.tipo) ? "Casa" : /terreno|lote/i.test(card.tipo) ? "Terreno" : "Apartamento";
  const aluguel = !card.venda && card.locacao > 0;
  const partes = [
    `${card.tipo} na ${card.rua}, ${card.bairro} — ${CIDADE}.`,
    [
      card.area && `${card.area} m² de área útil`,
      card.quartos && `${card.quartos} dormitório${card.quartos > 1 ? "s" : ""}${det.suites ? ` (sendo ${det.suites} suíte${det.suites > 1 ? "s" : ""})` : ""}`,
      card.vagas && `${card.vagas} vaga${card.vagas > 1 ? "s" : ""}`,
    ]
      .filter(Boolean)
      .join(", ") + ".",
    "Informações sujeitas a confirmação.",
  ];
  return {
    tipo,
    rua: card.rua,
    numero: "",
    bairro: card.bairro,
    cidade: CIDADE,
    estado: ESTADO,
    cep: "",
    pais: "Brasil",
    area: card.area,
    quartos: card.quartos,
    // O site não informa banheiros; cada suíte tem o seu, então usamos o
    // número de suítes como mínimo garantido.
    banheiros: det.suites,
    suites: det.suites,
    vagas: card.vagas,
    andar: 0,
    valor: aluguel ? card.locacao : card.venda,
    situacao: aluguel ? "Aluguel" : "Venda",
    disponivel: "true",
    descricao: `${partes.join("\n\n")}\n\nRef. parceiro: RTA-${card.slug}`,
  };
}

// ── API ────────────────────────────────────────────────────────────────────
async function login() {
  const { ADMIN_EMAIL: email, ADMIN_SENHA: senha } = process.env;
  if (!email || !senha) throw new Error("Defina ADMIN_EMAIL e ADMIN_SENHA");
  const r = await fetch(`${API}/login/usuario`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, senha }),
  });
  if (!r.ok) throw new Error(`login falhou (${r.status})`);
  const d = await r.json();
  return { token: typeof d.token === "string" ? d.token : d.token?.token, idPessoa: d.user?.id };
}

async function jaImportados() {
  const r = await fetch(`${API}/filtrar/imoveis`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ page: 1, limit: 5000, thumb: true, pais: "Brasil" }),
  });
  const d = r.ok ? await r.json() : [];
  const lista = Array.isArray(d) ? d : d.items || [];
  return new Set(lista.map((i) => String(i.descricao || "").match(/Ref\. parceiro: (\S+)/)?.[1]).filter(Boolean));
}

async function gravar(sessao, imovel, fotos) {
  const fd = new FormData();
  for (const [k, v] of Object.entries(imovel)) fd.append(k, String(v));
  if (sessao.idPessoa) fd.append("id_pessoa", String(sessao.idPessoa));
  for (const [i, url] of fotos.entries()) {
    const r = await fetch(url, { headers: { "User-Agent": UA } });
    if (r.ok) fd.append("imagem", new Blob([await r.blob()], { type: "image/jpeg" }), `foto-${i + 1}.jpg`);
  }
  const r = await fetch(`${API}/criar/imovel`, {
    method: "POST",
    headers: { Authorization: `Bearer ${sessao.token}` },
    body: fd,
  });
  if (!r.ok) throw new Error(`API respondeu ${r.status}: ${(await r.text()).slice(0, 200)}`);
  return (await r.json()).id;
}

// ── Execução ───────────────────────────────────────────────────────────────
const cards = lerCards(await baixarTexto(`${SITE}/`));
console.log(`Site RTA: ${cards.length} imóveis na página inicial.`);

const sessao = SIMULAR ? null : await login();
const existentes = SIMULAR ? new Set() : await jaImportados();
let gravados = 0;

for (const card of cards) {
  const ref = `RTA-${card.slug}`;
  if (existentes.has(ref)) {
    console.log(`  já importado: ${ref}`);
    continue;
  }
  await dormir(PAUSA_MS);
  const urlPagina = new URL(card.href, `${SITE}/`).href;
  let det;
  try {
    det = lerDetalhe(await baixarTexto(urlPagina), urlPagina);
  } catch (e) {
    console.log(`  pulei ${ref}: ${e.message}`);
    continue;
  }
  const imovel = montarImovel(card, det);
  const linha = `${imovel.tipo.padEnd(11)} | ${(imovel.rua + ", " + imovel.bairro).padEnd(38)} | R$ ${imovel.valor.toLocaleString("pt-BR").padStart(11)} | ${imovel.quartos}q ${imovel.suites}s ${imovel.vagas}v ${imovel.area}m² | ${det.fotos.length} fotos | ${ref}`;
  if (SIMULAR) {
    console.log("  [simulação] " + linha);
    continue;
  }
  try {
    const id = await gravar(sessao, imovel, det.fotos);
    console.log(`  OK #${id} ${linha}`);
    gravados++;
  } catch (e) {
    console.log(`  ERRO ${ref}: ${e.message}`);
  }
}

console.log(`\nConcluído: ${gravados} imóveis gravados${SIMULAR ? " (simulação)" : ""}.`);
