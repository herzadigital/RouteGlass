package httpapi

import (
	"context"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/herzadigital/routeglass/internal/config"
	"github.com/herzadigital/routeglass/internal/diagnostics"
	"github.com/herzadigital/routeglass/internal/juniper"
	"github.com/herzadigital/routeglass/internal/mikrotik"
)

type pageData struct {
	config.Config
	VisitorIP   string
	BGPEnabled  bool
	BGPProvider string
}

const page = `<!doctype html>
<html><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>RouteGlass - {{.CompanyName}}</title>
<style>
body{font-family:Arial,sans-serif;background:#f5f7fa;color:#222;margin:0}
main{max-width:980px;margin:35px auto;background:#fff;padding:30px;border-radius:12px;box-shadow:0 2px 15px #0002}
.info{margin:0 0 26px;line-height:1.65}.info div{display:block}
input,select,button{padding:10px;margin:4px;font-size:14px}
pre{white-space:pre-wrap;background:#111;color:#eee;padding:16px;border-radius:8px;min-height:160px}
section{margin-top:28px;padding-top:20px;border-top:1px solid #ddd}
footer{display:flex;justify-content:space-between;gap:20px;flex-wrap:wrap;margin-top:30px;padding-top:20px;border-top:1px solid #ddd;font-size:13px;color:#666}
a{color:#1769aa;text-decoration:none}.test-files a{margin-right:10px}
.result-card{display:none;margin-top:16px;border:1px solid #ddd;border-radius:8px;padding:16px;background:#fafafa}
.result-grid{display:grid;grid-template-columns:220px 1fr;gap:8px 16px}.label{font-weight:bold}
</style></head>
<body><main>
<h1>{{.CompanyName}} RouteGlass</h1>

<div class="info">
<div><b>Location:</b> {{.Location}}</div>
<div><b>ASN:</b> AS{{.ASN}}</div>
{{if .TestIPv4}}<div><b>Test IPv4:</b> {{.TestIPv4}}</div>{{end}}
{{if .TestIPv6}}<div><b>Test IPv6:</b> {{.TestIPv6}}</div>{{end}}
{{if .TestFiles}}<div class="test-files"><b>Test Files:</b> {{range .TestFiles}}<a href="/testfiles/{{.}}MB.bin">{{.}}MB</a>{{end}}</div>{{end}}
<div><b>Your IP Address:</b> {{.VisitorIP}}</div>
</div>

<section>
<h2>Network Diagnostics</h2>
<form id="f">
<select name="type" id="type">
<option value="ping">Ping</option>
<option value="mtr">MTR</option>
<option value="traceroute">Traceroute</option>
{{if .BGPEnabled}}<option value="bgp">BGP</option>{{end}}
</select>
<input name="target" placeholder="IP address or hostname" required>
<button>Run Test</button>
</form>

<div id="bgpResult" class="result-card">
<h3>Route Lookup Result</h3>
<div class="result-grid" id="bgpFields"></div>
</div>
<pre id="o">Ready.</pre>
</section>

<footer>
<div>©2019-2025. {{.CompanyName}} All Rights Reserved.<br>
Developed by <a href="https://herza.id" target="_blank" rel="noopener">PT Herza Digital Indonesia</a></div>
<div>Info:
<a href="https://stat.ripe.net/AS{{.ASN}}" target="_blank">RIPE</a> |
<a href="https://asrank.caida.org/asns?asn={{.ASN}}" target="_blank">ASRank</a> |
<a href="https://radar.qrator.net/as{{.ASN}}" target="_blank">RADAR</a> |
<a href="https://as{{.ASN}}.peeringdb.com" target="_blank">PeeringDB</a> |
<a href="https://bgp.he.net/AS{{.ASN}}" target="_blank">HE.net</a>
</div>
</footer>

</main>
<script>
const f=document.getElementById("f"),o=document.getElementById("o"),bgpResult=document.getElementById("bgpResult"),bgpFields=document.getElementById("bgpFields");
function esc(v){const d=document.createElement("div");d.textContent=v??"";return d.innerHTML}
function row(label,value){return '<div class="label">'+esc(label)+'</div><div>'+esc(value||"-")+'</div>'}
f.addEventListener("submit",async e=>{
 e.preventDefault();
 const form=new FormData(f),type=form.get("type");
 o.style.display="block";bgpResult.style.display="none";o.textContent=type==="bgp"?"Looking up route...":"Running...";
 try{
   const r=await fetch("/api/diagnostics?"+new URLSearchParams(form));
   const d=await r.json();
   if(!r.ok||d.error){o.textContent=d.error||"Request failed";return}
   if(type==="bgp"){
     o.style.display="none";
     bgpFields.innerHTML=
       row("Target IP",d.target)+
       row("Best Matching Prefix",d.prefix)+
       row("Next Hop",d.gateway)+
       row("BGP AS Path",d.as_path)+
       row("Protocol",d.protocol)+
       row("Administrative Distance",d.distance);
     bgpResult.style.display="block";
   }else{o.textContent=d.output||"No output"}
 }catch(x){o.textContent="Request failed: "+x}
});
</script>
</body></html>`

func visitorIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	h, _, e := net.SplitHostPort(r.RemoteAddr)
	if e == nil {
		return h
	}
	return r.RemoteAddr
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func New(c config.Config) http.Handler {
	t := template.Must(template.New("p").Parse(page))
	m := http.NewServeMux()

	var mk *mikrotik.Client
	var jn *juniper.Client
	provider := ""
	if c.MikroTik.Enabled {
		mk = mikrotik.New(c.MikroTik)
		provider = "MikroTik"
	}
	if c.Juniper.Enabled {
		jn = juniper.New(c.Juniper)
		provider = "Juniper"
	}

	lookupBGP := func(r *http.Request, target string) (mikrotik.Result, error) {
		if mk == nil && jn == nil {
			return mikrotik.Result{}, mikrotik.ErrDisabled
		}
		if jn != nil {
			timeout := time.Duration(c.Juniper.TimeoutSeconds) * time.Second
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			return jn.Lookup(ctx, target)
		}
		timeout := time.Duration(c.MikroTik.TimeoutSeconds) * time.Second
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		return mk.Lookup(ctx, target)
	}

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_ = t.Execute(w, pageData{
			Config: c, VisitorIP: visitorIP(r), BGPEnabled: mk != nil || jn != nil, BGPProvider: provider,
		})
	})

	m.HandleFunc("/api/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		typ := strings.TrimSpace(r.URL.Query().Get("type"))
		target := strings.TrimSpace(r.URL.Query().Get("target"))

		if typ == "bgp" {
			res, err := lookupBGP(r, target)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, res)
			return
		}

		out, err := diagnostics.Run(typ, target)
		if err != nil && out == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"output": out})
	})

	// Backward-compatible endpoint retained for API consumers.
	m.HandleFunc("/api/bgp", func(w http.ResponseWriter, r *http.Request) {
		res, err := lookupBGP(r, strings.TrimSpace(r.URL.Query().Get("target")))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})

	return m
}
