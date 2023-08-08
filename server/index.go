package server

var indexHTMLHeader = []byte(`<!doctype html>
<html>
<body onload="main();">
`)

var indexHTMLFooter = []byte(`
<p><a href="/quit">Quit</a></p>
<div style="width: 90vw; height: 90vh;">
	<img id="latest" style="max-width: 100%; max-height: 100%; height: auto;">
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
