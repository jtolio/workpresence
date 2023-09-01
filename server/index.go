package server

var indexHTMLHeader = []byte(`<!doctype html>
<html>
<head>
<meta charset=utf-8>
<meta name=viewport content="width=device-width, initial-scale=1.0">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
</head>
<body onload="main();">
`)

var indexHTMLFooter = []byte(`
<p><a href="/quit">Quit</a></p>
<div style="width: 90vw; height: 90vh;">
	<img id="latest" style="max-width: 100%; min-width: 400px; max-height: 100%; height: auto;">
</div>

<script>
	function tick() {
		document.getElementById("latest").src = "./latest#" + new Date().getTime();
	}
	function main() {
		setInterval(tick, 1000);
		tick();
	}
</script>
</body>
</html>
`)
