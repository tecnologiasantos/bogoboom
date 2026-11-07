(function () {
  "use strict";

  var T = {
    en: {
      nav_how: "How it works", nav_why: "Why not hotels", nav_launch: "Bogotá first", nav_join: "Join the waitlist",
      hero_title: "Your smart start, anywhere.",
      hero_sub: "A worldwide company that connects real estate agencies, HR departments and newcomers.",
      cta_join: "Join the waitlist", cta_how: "See how it works",
      panel_hint: "Tap a house to see who Bogoboom is for.",
      aud_new_title: "Newcomers", aud_new_lead: "Find a verified long-term home before you land. It's free.",
      aud_new_1: "Browse listings from vetted agencies", aud_new_2: "Filter by budget, neighborhood and lifestyle", aud_new_3: "Avoid scams and settle in faster", aud_new_cta: "Find a home",
      aud_hr_title: "HR teams", aud_hr_lead: "Relocate talent into real homes, not hotel rooms.",
      aud_hr_1: "Cut hotel and short-term housing costs", aud_hr_2: "Give new hires a smoother first month", aud_hr_3: "Follow every relocation in one place", aud_hr_cta: "Plan a relocation",
      aud_ag_title: "Real estate agencies", aud_ag_lead: "Meet renters who are looking for a long-term home.",
      aud_ag_1: "Get matched with professionals, students and expats", aud_ag_2: "Fill vacancies with verified leads", aud_ag_3: "Sign up free, subscribe when you're ready to scale", aud_ag_cta: "List your properties",
      why_title: "Hotels are the expensive way to start.",
      why_body: "Companies that hire from other cities often put new hires in hotels. It inflates relocation budgets, delays integration and stresses people who are still learning a new city.",
      why_before: "With a hotel stay", why_b1: "You pay per night while the housing search drags on", why_b2: "New hires feel like visitors, not residents", why_b3: "HR coordinates by email and spreadsheet",
      why_after: "With Bogoboom", why_a1: "A real home is lined up before arrival", why_a2: "New hires start their job already settled", why_a3: "HR follows progress in one place",
      how_title: "From offer letter to front door.",
      how_1_t: "Say who is moving", how_1_p: "Tell us the city, the dates and the budget. Newcomers, HR teams and agencies each start from their own door.",
      how_2_t: "Verified agencies respond", how_2_p: "Only vetted real estate agencies show homes, so the options are real.",
      how_3_t: "Compare matching homes", how_3_p: "Sort by budget, neighborhood and move-in date, then shortlist your favorites.",
      how_4_t: "Sign before you land", how_4_p: "Secure long-term housing ahead of arrival and skip the hotel nights.",
      launch_title: "Launching in Bogotá. Built for anywhere.",
      launch_body: "We start in Colombia's capital, where BPO, tech and education teams hire people from across the country and the world. The model is designed to repeat in every growing city.",
      join_title: "Be first in line.", join_body: "Tell us which side you're on and we'll write when Bogotá opens.",
      f_name: "Your name", f_email: "Email", f_role: "I am a", f_r_new: "Newcomer", f_r_hr: "HR team", f_r_ag: "Real estate agency",
      foot: "Bogoboom and the MIRA job portal are technical properties of Tecnologia Santos.",
      mail_subject: "Bogoboom waitlist"
    },
    es: {
      nav_how: "Cómo funciona", nav_why: "Por qué no hoteles", nav_launch: "Primero Bogotá", nav_join: "Unirme a la lista",
      hero_title: "Tu comienzo inteligente, en cualquier lugar.",
      hero_sub: "Una empresa global que conecta inmobiliarias, equipos de RR. HH. y recién llegados.",
      cta_join: "Unirme a la lista", cta_how: "Ver cómo funciona",
      panel_hint: "Toca una casa para ver para quién es Bogoboom.",
      aud_new_title: "Recién llegados", aud_new_lead: "Encuentra una vivienda verificada a largo plazo antes de aterrizar. Es gratis.",
      aud_new_1: "Explora anuncios de inmobiliarias verificadas", aud_new_2: "Filtra por presupuesto, barrio y estilo de vida", aud_new_3: "Evita estafas e instálate más rápido", aud_new_cta: "Buscar vivienda",
      aud_hr_title: "Equipos de RR. HH.", aud_hr_lead: "Traslada talento a hogares reales, no a habitaciones de hotel.",
      aud_hr_1: "Reduce los costos de hotel y alojamiento temporal", aud_hr_2: "Ofrece un primer mes más tranquilo a los nuevos empleados", aud_hr_3: "Sigue cada traslado en un solo lugar", aud_hr_cta: "Planear un traslado",
      aud_ag_title: "Inmobiliarias", aud_ag_lead: "Conoce inquilinos que buscan un hogar a largo plazo.",
      aud_ag_1: "Conecta con profesionales, estudiantes y expatriados", aud_ag_2: "Ocupa tus vacantes con contactos verificados", aud_ag_3: "Regístrate gratis y suscríbete cuando quieras crecer", aud_ag_cta: "Publicar propiedades",
      why_title: "El hotel es la forma cara de empezar.",
      why_body: "Las empresas que contratan en otras ciudades suelen alojar a los nuevos empleados en hoteles. Eso infla el presupuesto de traslado, retrasa la integración y estresa a quien aún está conociendo una ciudad nueva.",
      why_before: "Con hotel", why_b1: "Pagas por noche mientras la búsqueda de vivienda se alarga", why_b2: "Los nuevos empleados se sienten visitantes, no residentes", why_b3: "RR. HH. coordina por correo y hojas de cálculo",
      why_after: "Con Bogoboom", why_a1: "Un hogar real listo antes de la llegada", why_a2: "Los nuevos empleados empiezan ya instalados", why_a3: "RR. HH. sigue el avance en un solo lugar",
      how_title: "De la carta de oferta a la puerta de casa.",
      how_1_t: "Cuéntanos quién se muda", how_1_p: "Indica la ciudad, las fechas y el presupuesto. Recién llegados, RR. HH. e inmobiliarias empiezan cada uno por su puerta.",
      how_2_t: "Responden inmobiliarias verificadas", how_2_p: "Solo inmobiliarias verificadas publican viviendas, así que las opciones son reales.",
      how_3_t: "Compara viviendas ideales", how_3_p: "Ordena por presupuesto, barrio y fecha de mudanza, y guarda tus favoritas.",
      how_4_t: "Firma antes de aterrizar", how_4_p: "Asegura vivienda a largo plazo antes de llegar y olvida las noches de hotel.",
      launch_title: "Lanzamos en Bogotá. Hecho para cualquier lugar.",
      launch_body: "Empezamos en la capital de Colombia, donde equipos de BPO, tecnología y educación contratan personas de todo el país y del mundo. El modelo está pensado para repetirse en cada ciudad en crecimiento.",
      join_title: "Sé de los primeros.", join_body: "Dinos de qué lado estás y te escribiremos cuando abra Bogotá.",
      f_name: "Tu nombre", f_email: "Correo electrónico", f_role: "Soy", f_r_new: "Recién llegado", f_r_hr: "Equipo de RR. HH.", f_r_ag: "Inmobiliaria",
      foot: "Bogoboom y el portal de empleo MIRA son propiedades técnicas de Tecnologia Santos.",
      mail_subject: "Lista de espera Bogoboom"
    },
    pt: {
      nav_how: "Como funciona", nav_why: "Por que não hotel", nav_launch: "Bogotá primeiro", nav_join: "Entrar na lista",
      hero_title: "Seu começo inteligente, em qualquer lugar.",
      hero_sub: "Uma empresa global que conecta imobiliárias, equipes de RH e recém-chegados.",
      cta_join: "Entrar na lista de espera", cta_how: "Ver como funciona",
      panel_hint: "Toque numa casa para ver para quem é o Bogoboom.",
      aud_new_title: "Recém-chegados", aud_new_lead: "Encontre uma moradia verificada de longo prazo antes de chegar. É grátis.",
      aud_new_1: "Veja anúncios de imobiliárias verificadas", aud_new_2: "Filtre por orçamento, bairro e estilo de vida", aud_new_3: "Evite golpes e se instale mais rápido", aud_new_cta: "Encontrar moradia",
      aud_hr_title: "Equipes de RH", aud_hr_lead: "Leve talentos para casas de verdade, não para quartos de hotel.",
      aud_hr_1: "Reduza os custos com hotel e hospedagem temporária", aud_hr_2: "Dê um primeiro mês mais tranquilo aos novos contratados", aud_hr_3: "Acompanhe cada mudança em um só lugar", aud_hr_cta: "Planejar uma mudança",
      aud_ag_title: "Imobiliárias", aud_ag_lead: "Conheça inquilinos que procuram um lar de longo prazo.",
      aud_ag_1: "Conecte-se com profissionais, estudantes e expatriados", aud_ag_2: "Preencha vagas com contatos verificados", aud_ag_3: "Cadastre-se grátis e assine quando quiser crescer", aud_ag_cta: "Anunciar imóveis",
      why_title: "Hotel é a forma cara de começar.",
      why_body: "Empresas que contratam em outras cidades costumam hospedar os novos contratados em hotéis. Isso infla o orçamento de mudança, atrasa a integração e estressa quem ainda está conhecendo uma cidade nova.",
      why_before: "Com hotel", why_b1: "Você paga por noite enquanto a busca por moradia se arrasta", why_b2: "Os novos contratados se sentem visitantes, não moradores", why_b3: "O RH coordena por e-mail e planilha",
      why_after: "Com o Bogoboom", why_a1: "Uma casa de verdade garantida antes da chegada", why_a2: "Os novos contratados começam o trabalho já instalados", why_a3: "O RH acompanha o andamento em um só lugar",
      how_title: "Da carta-oferta à porta de casa.",
      how_1_t: "Diga quem vai se mudar", how_1_p: "Informe a cidade, as datas e o orçamento. Recém-chegados, RH e imobiliárias começam cada um pela sua porta.",
      how_2_t: "Imobiliárias verificadas respondem", how_2_p: "Só imobiliárias verificadas mostram imóveis, então as opções são reais.",
      how_3_t: "Compare as casas ideais", how_3_p: "Ordene por orçamento, bairro e data de mudança e salve suas favoritas.",
      how_4_t: "Assine antes de chegar", how_4_p: "Garanta moradia de longo prazo antes da chegada e pule as noites de hotel.",
      launch_title: "Começando em Bogotá. Feito para qualquer lugar.",
      launch_body: "Começamos na capital da Colômbia, onde equipes de BPO, tecnologia e educação contratam pessoas de todo o país e do mundo. O modelo foi pensado para se repetir em cada cidade em crescimento.",
      join_title: "Seja dos primeiros.", join_body: "Conte de que lado você está e escreveremos quando Bogotá abrir.",
      f_name: "Seu nome", f_email: "E-mail", f_role: "Eu sou", f_r_new: "Recém-chegado", f_r_hr: "Equipe de RH", f_r_ag: "Imobiliária",
      foot: "Bogoboom e o portal de vagas MIRA são propriedades técnicas da Tecnologia Santos.",
      mail_subject: "Lista de espera Bogoboom"
    }
  };

  var lang = "en";
  try { lang = localStorage.getItem("bogoboom-lang") || ""; } catch (e) {}
  if (!T[lang]) {
    var nav = (navigator.language || "en").slice(0, 2).toLowerCase();
    lang = T[nav] ? nav : "en";
  }

  function applyLang(l) {
    lang = l;
    document.documentElement.lang = l;
    document.querySelectorAll("[data-i18n]").forEach(function (el) {
      var v = T[l][el.getAttribute("data-i18n")];
      if (v) el.textContent = v;
    });
    document.querySelectorAll("[data-lang]").forEach(function (b) {
      b.setAttribute("aria-pressed", b.getAttribute("data-lang") === l ? "true" : "false");
    });
    var labels = { newcomers: "aud_new_title", hr: "aud_hr_title", agencies: "aud_ag_title" };
    document.querySelectorAll(".bld").forEach(function (g) {
      g.setAttribute("aria-label", T[l][labels[g.getAttribute("data-aud")]]);
    });
    try { localStorage.setItem("bogoboom-lang", l); } catch (e) {}
  }

  document.querySelectorAll("[data-lang]").forEach(function (b) {
    b.addEventListener("click", function () { applyLang(b.getAttribute("data-lang")); });
  });

  // houses drive the audience panel
  var houses = document.querySelectorAll(".bld");
  var bodies = document.querySelectorAll("[data-aud-panel]");
  function select(aud) {
    houses.forEach(function (g) { g.setAttribute("aria-pressed", g.getAttribute("data-aud") === aud ? "true" : "false"); });
    bodies.forEach(function (p) { p.hidden = p.getAttribute("data-aud-panel") !== aud; });
  }
  houses.forEach(function (g) {
    g.addEventListener("click", function () { select(g.getAttribute("data-aud")); });
    g.addEventListener("keydown", function (e) {
      if (e.key === "Enter" || e.key === " ") { e.preventDefault(); select(g.getAttribute("data-aud")); }
    });
  });

  // panel buttons pre-select the role in the form
  document.querySelectorAll("[data-role]").forEach(function (a) {
    a.addEventListener("click", function () {
      var sel = document.getElementById("role");
      if (sel) sel.value = a.getAttribute("data-role");
    });
  });

  // no backend yet: the form opens the visitor's email app with the details filled in
  document.getElementById("waitlist").addEventListener("submit", function (e) {
    e.preventDefault();
    var f = e.target;
    var role = f.role.options[f.role.selectedIndex].textContent;
    var body = "Name: " + f.name.value + "\nEmail: " + f.email.value + "\nRole: " + role + "\nLanguage: " + lang;
    window.location.href = "mailto:support@tecnologiasantos.com?subject=" +
      encodeURIComponent(T[lang].mail_subject) + "&body=" + encodeURIComponent(body);
  });

  applyLang(lang);
})();
