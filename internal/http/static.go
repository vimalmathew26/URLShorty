package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterStatic wires a tiny inline HTML page at GET "/".
func RegisterStatic(r *gin.Engine) {
	const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>urlshorty</title>
<style>
body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,Cantarell,Noto Sans,sans-serif;margin:0;padding:2rem;background:#0b0b0c;color:#e8e8ea}
.container{max-width:680px;margin:0 auto}
.card{background:#151517;border:1px solid #2b2b2f;border-radius:12px;padding:1.25rem}
h1{font-size:1.25rem;margin:0 0 1rem}
input,button{font-size:1rem}
input[type=text]{width:100%;padding:.75rem;border-radius:8px;border:1px solid #2b2b2f;background:#0f0f11;color:#e8e8ea}
.row{display:flex;gap:.5rem;margin-top:.75rem}
.row button{padding:.75rem 1rem;border:1px solid #2b2b2f;background:#1f1f23;color:#e8e8ea;border-radius:8px;cursor:pointer}
small{opacity:.7}
pre{white-space:pre-wrap;word-break:break-word;background:#0f0f11;border:1px solid #2b2b2f;border-radius:8px;padding:.75rem}
a{color:#97b3ff}
</style>
</head>
<body>
<div class="container">
  <div class="card">
    <h1>urlshorty — make a short link</h1>
    <input id="url" type="text" placeholder="https://example.com/very/long/link"/>
    <div class="row">
      <input id="custom" type="text" placeholder="custom alias (optional)"/>
      <button id="go">Shorten</button>
    </div>
    <small>Optional ISO-8601 expiry (e.g., 2025-12-31T23:59:59Z)</small>
    <input id="exp" type="text" placeholder="expires_at (optional)"/>
    <div id="out" style="margin-top:1rem"></div>
  </div>
  <p style="opacity:.7;margin-top:1rem">API: <code>POST /api/shorten</code>, <code>GET /:code</code>, <code>GET /api/:code</code></p>
</div>
<script>
async function shorten(){
  const url = document.getElementById('url').value.trim();
  const custom = document.getElementById('custom').value.trim();
  const exp = document.getElementById('exp').value.trim();
  const body = { url };
  if(custom) body.custom = custom;
  if(exp) body.expires_at = exp;
  const res = await fetch('/api/shorten', {
    method:'POST',
    headers:{'Content-Type':'application/json'},
    body:JSON.stringify(body)
  });
  const out = document.getElementById('out');
  const data = await res.json().catch(()=>({}));
  if(!res.ok){ out.innerHTML = '<pre>'+JSON.stringify(data,null,2)+'</pre>'; return; }
  out.innerHTML = '<pre>'+JSON.stringify(data,null,2)+'</pre>'+
    '<p><a target="_blank" rel="noopener" href="'+data.short_url+'">'+data.short_url+'</a></p>';
}
document.getElementById('go').addEventListener('click', shorten);
document.getElementById('url').addEventListener('keydown', e=>{ if(e.key==='Enter') shorten(); });
</script>
</body>
</html>`
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
	})
}
