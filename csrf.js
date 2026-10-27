/* Adds a per-session CSRF token to every POST form (static HTML pages cannot render PHP tokens).
   The token is fetched from /backend/profiles/csrf.php (same origin, session cookie). */
(function () {
  function forms() { return Array.prototype.filter.call(document.forms, function (f) { return (f.method || '').toLowerCase() === 'post'; }); }
  function fill(token) {
    forms().forEach(function (f) {
      var i = f.querySelector('input[name="csrf_token"]');
      if (!i) { i = document.createElement('input'); i.type = 'hidden'; i.name = 'csrf_token'; f.appendChild(i); }
      i.value = token;
    });
  }
  function load() {
    fetch('/backend/profiles/csrf.php', { credentials: 'same-origin', cache: 'no-store' })
      .then(function (r) { return r.json(); })
      .then(function (d) { fill(d.token); })
      .catch(function () { /* the server will reject the submit; the user can reload */ });
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', load); else load();
  window.addEventListener('pageshow', function (e) { if (e.persisted) load(); }); // back/forward cache
})();
