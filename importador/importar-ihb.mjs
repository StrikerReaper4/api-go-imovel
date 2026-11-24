// Importa imóveis PÚBLICOS do site parceiro (IHB Brazil) para a nossa API.
//
// Lê só as páginas públicas, como um visitante (o robots.txt deles libera
// as páginas e proíbe /api/ — que nunca é acessado aqui). Os dados vêm do
// bloco schema.org (JSON-LD) que o próprio site publica para o Google.
//
// Uso (com a API local no ar):
//   node importador/importar-ihb.mjs --eua 5 --portugal 5
//   node importador/importar-ihb.mjs --todos          (todos os imóveis; pula os já importados)
//   node importador/importar-ihb.mjs --eua 5 --portugal 5 --simular   (não grava)
//
// Variáveis: API_URL (padrão http://localhost:8080), ADMIN_EMAIL, ADMIN_SENHA.

const SITE = "https://www.ihbbrazil.com.br";
const API = process.env.API_URL || "http://localhost:8080";
const MAX_FOTOS = 12;
const PAUSA_MS = 1500; // educação com o servidor deles: uma página por vez

const args = process.argv.slice(2);
const opt = (nome, padrao) => {
  const i = args.indexOf(`--${nome}`);
  return i >= 0 ? Number(args[i + 1]) : padrao;
};
const TODOS = args.includes("--todos");
const QTD_EUA = TODOS ? Infinity : opt("eua", 5);
const QTD_PT = TODOS ? Infinity : opt("portugal", 5);
const SIMULAR = args.includes("--simular");

const UA = "BretasImoveis-Importador/1.0 (parceria IHB; contato via site bretasimoveis)";
const dormir = (ms) => new Promise((r) => setTimeout(r, ms));

async function baixarTexto(url) {
  const r = await fetch(url, { headers: { "User-Agent": UA } });
  if (!r.ok) throw new Error(`${r.status} em ${url}`);
  return r.text();
}

const decodificar = (s) =>
  s
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&nbsp;/g, " ");

// ── Leitura de uma página de imóvel ────────────────────────────────────────
function lerImovel(html, url) {
  const blocos = [...html.matchAll(/<script[^>]*application\/ld\+json[^>]*>([\s\S]*?)<\/script>/g)]
    .map((m) => {
      try {
        return JSON.parse(m[1]);
      } catch {
        return null;
      }
    })
    .filter(Boolean);
  const anuncio = blocos.find((b) => b["@type"] === "RealEstateListing");
  if (!anuncio) throw new Error("página sem dados estruturados");

  const ent = anuncio.mainEntity || {};
  const idParceiro = url.match(/\/(\d+)$/)?.[1];

  // Descrição completa (a do JSON-LD vem cortada com "…").
  const desc = html.match(/<p class="pre-line[^"]*">([\s\S]*?)<\/p>/);
  const descricao = decodificar((desc ? desc[1] : anuncio.description || "").replace(/<br\s*\/?>/g, "\n").replace(/<[^>]+>/g, ""))
    .replace(/\n{3,}/g, "\n\n")
    .trim();

  // Linha de endereço exibida no topo: "Bairro, Cidade, País".
  const loc = html.match(/lucide-map-pin[\s\S]*?<\/svg>([^<]+)</);
  const partes = loc ? decodificar(loc[1]).split(",").map((s) => s.trim()) : [];

  const paisIso = ent.address?.addressCountry;
  const pais = paisIso === "PT" ? "Portugal" : paisIso === "US" ? "Estados Unidos" : partes.at(-1) || "";
  const cidade = ent.address?.addressLocality || partes.at(-2) || "";
  const bairro = partes.length >= 3 ? partes[0] : "";

  const tipoLd = ent["@type"] || "";
  const nome = anuncio.name || "";
  const tipo = /terreno|lote/i.test(nome)
    ? "Terreno"
    : tipoLd === "SingleFamilyResidence" || /^casa|casa à venda|\bcasa\b/i.test(nome)
    ? "Casa"
    : "Apartamento";

  const oferta = anuncio.offers || {};
  const valor = Math.round(Number(oferta.price ?? oferta.lowPrice ?? 0));

  // Fotos: as versões "_o" (1280px) da pasta deste imóvel, na ordem da página.
  const fotos = [];
  const re = new RegExp(`https://cdn\\.vistahost\\.com\\.br/[^"'\\s)]*?/fotos/${idParceiro}/[^"'\\s)?]+_o\\.jpe?g`, "gi");
  for (const m of html.matchAll(re)) if (!fotos.includes(m[0])) fotos.push(m[0]);
  for (const f of anuncio.image || []) if (!fotos.includes(f)) fotos.push(f);

  const area = ent.floorSize?.value ? Math.round(Number(ent.floorSize.value)) : 0;

  return {
    origem: url,
    empreendimento: tipoLd === "ApartmentComplex" || /^empreendimento/i.test(nome),
    codigo: anuncio.identifier || idParceiro,
    nome,
    imovel: {
      tipo,
      rua: "",
      numero: "",
      bairro,
      cidade,
      estado: paisIso === "US" ? "FL" : "",
      cep: "",
      pais,
      area,
      quartos: Number(ent.numberOfBedrooms ?? 0) || 0,
      banheiros: Number(ent.numberOfBathroomsTotal ?? 0) || 0,
      suites: 0,
      vagas: 0,
      andar: 0,
      valor,
      situacao: "Venda",
      disponivel: "true",
      descricao: `${nome}\n\n${descricao}\n\nRef. parceiro: ${anuncio.identifier || "IHB" + idParceiro}`.trim(),
    },
    fotos: fotos.slice(0, MAX_FOTOS),
    moeda: oferta.priceCurrency,
  };
}

// ── API ────────────────────────────────────────────────────────────────────
async function login() {
  const email = process.env.ADMIN_EMAIL;
  const senha = process.env.ADMIN_SENHA;
  if (!email || !senha) throw new Error("Defina ADMIN_EMAIL e ADMIN_SENHA");
  const r = await fetch(`${API}/login/usuario`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, senha }),
  });
  if (!r.ok) throw new Error(`login falhou (${r.status})`);
  const d = await r.json();
  const token = typeof d.token === "string" ? d.token : d.token?.token;
  return { token, idPessoa: d.user?.id };
}

async function jaImportados() {
  const r = await fetch(`${API}/filtrar/imoveis`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ page: 1, limit: 5000, thumb: true }),
  });
  const lista = r.ok ? await r.json() : [];
  const refs = new Set();
  for (const i of Array.isArray(lista) ? lista : lista.items || []) {
    const m = String(i.descricao || "").match(/Ref\. parceiro: (\S+)/);
    if (m) refs.add(m[1]);
  }
  return refs;
}

async function gravar(sessao, dados) {
  const fd = new FormData();
  for (const [k, v] of Object.entries(dados.imovel)) fd.append(k, String(v));
  if (sessao.idPessoa) fd.append("id_pessoa", String(sessao.idPessoa));
  for (const [i, url] of dados.fotos.entries()) {
    const r = await fetch(url, { headers: { "User-Agent": UA } });
    if (!r.ok) continue;
    const blob = await r.blob();
    fd.append("imagem", new Blob([blob], { type: "image/jpeg" }), `foto-${i + 1}.jpg`);
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
const PT = /(lisboa|porto|portugal|algarve|cascais|braga|gaia|matosinhos|coimbra|aveiro|sintra|oeiras|faro|lagos|albufeira|setubal|setúbal|madeira|funchal|vilamoura|comporta|ericeira|estoril)/i;
const EUA = /(miami|orlando|kissimmee|clermont|davenport|winter-garden|florida|sunny-isles|fort-lauderdale|brickell|doral|aventura|hollywood|tampa|celebration)/i;

const sitemap = await baixarTexto(`${SITE}/sitemap.xml`);
const urls = [...sitemap.matchAll(/<loc>([^<]*\/imovel\/[^<]*)<\/loc>/g)].map((m) => m[1]);
// Ordena pelos que o link já indica o país e deixa os sem cidade reconhecível
// por último — mas o país que vale é o lido na página, não o do link.
const prioridade = (u) => (PT.test(u) || EUA.test(u) ? 0 : 1);
urls.sort((a, b) => prioridade(a) - prioridade(b));
console.log(`Sitemap: ${urls.length} imóveis. Alvo: EUA ${QTD_EUA === Infinity ? "todos" : QTD_EUA}, Portugal ${QTD_PT === Infinity ? "todos" : QTD_PT}.`);

const alvo = { "Estados Unidos": QTD_EUA, Portugal: QTD_PT };
const feitos = { "Estados Unidos": 0, Portugal: 0 };
const pulados = { empreendimento: 0, "sem preço": 0, "já importado": 0, "outro país": 0, "erro de leitura": 0, "erro ao gravar": 0 };

const sessao = SIMULAR ? null : await login();
const existentes = SIMULAR ? new Set() : await jaImportados();
const resumo = [];
const inicio = Date.now();

for (const [n, url] of urls.entries()) {
  if (feitos["Estados Unidos"] >= alvo["Estados Unidos"] && feitos.Portugal >= alvo.Portugal) break;
  await dormir(PAUSA_MS);
  let dados;
  try {
    dados = lerImovel(await baixarTexto(url), url);
  } catch (e) {
    pulados["erro de leitura"]++;
    console.log(`  pulei ${url}: ${e.message}`);
    continue;
  }
  const pais = dados.imovel.pais;
  if (!(pais in alvo)) { pulados["outro país"]++; continue; }
  if (feitos[pais] >= alvo[pais]) continue;
  if (dados.imovel.valor <= 0) { pulados["sem preço"]++; continue; }
  // Empreendimentos (condomínio com várias unidades) não trazem quartos,
  // banheiros nem área, e as unidades deles já aparecem como anúncios
  // próprios. Ficam de fora para não duplicar nem gravar ficha vazia.
  if (dados.empreendimento) { pulados.empreendimento++; continue; }
  if (existentes.has(String(dados.codigo))) { pulados["já importado"]++; continue; }

  const linha = `${pais.padEnd(14)} | ${dados.imovel.tipo.padEnd(11)} | ${dados.imovel.cidade.padEnd(16)} | ${dados.moeda} ${dados.imovel.valor.toLocaleString("pt-BR").padStart(12)} | ${dados.imovel.quartos}q ${dados.imovel.banheiros}b ${dados.imovel.area}m² | ${dados.fotos.length} fotos | ${dados.codigo}`;
  if (SIMULAR) {
    console.log("  [simulação] " + linha);
  } else {
    try {
      const id = await gravar(sessao, dados);
      existentes.add(String(dados.codigo));
      console.log(`  [${n + 1}/${urls.length}] OK #${id} ` + linha);
      resumo.push({ id, pais });
    } catch (e) {
      pulados["erro ao gravar"]++;
      console.log(`  ERRO ${dados.codigo}: ${e.message}`);
      continue;
    }
  }
  feitos[pais]++;
}

const min = ((Date.now() - inicio) / 60000).toFixed(1);
console.log(`\nConcluído em ${min} min: ${resumo.length} imóveis gravados${SIMULAR ? " (simulação)" : ""}.`);
console.log(`  EUA: ${feitos["Estados Unidos"]} | Portugal: ${feitos.Portugal}`);
console.log(`  Pulados: ${Object.entries(pulados).map(([k, v]) => `${k} ${v}`).join(" | ")}`);
