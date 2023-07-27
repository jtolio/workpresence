package storage

var indexHTML = []byte(`<!doctype html>
<html>
<body onload="main();">
<div style="width: 90vw; height: 90vh;">
	<img id="latest" style="max-width: 100%; max-height: 100%; height: auto;">
</div>

<script>
	async function tick() {
		var response = await fetch("./latest");
		if (response.status != 200) {
			if (response.status != 404) {
				console.log(response);
			}
			return;
		}
		document.getElementById("latest").src = await response.text()
	}
	async function main() {
		setInterval(tick, 5000);
		tick();
	}
</script>
</body>
</html>
`)
