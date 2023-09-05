package storage

import (
	"html/template"
)

var indexHTML = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head>
<meta charset=utf-8>
<meta name=viewport content="width=device-width, initial-scale=1.0">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
<title>Work Presence</title>
</head>
<body onload="main();">
  <ul id="bucket-content"></ul>
  <p style="font-size:smaller; text-align: right;">
    This page is served from <a href="?wrap=1">all around the world</a>.</p>

  <script src="https://sdk.amazonaws.com/js/aws-sdk-2.1168.0.min.js"></script>
  <script>
    AWS.config.update({
      accessKeyId: '{{.Config.ListingAccessKey}}',
      secretAccessKey: '{{.Config.ListingSecretKey}}',
      region: 'global',
      endpoint: 'https://gateway.storjshare.io',
      s3ForcePathStyle: true
    });

    const s3 = new AWS.S3();

    const bucketName = '{{.Config.Bucket}}';
    const bucketPrefix = '{{.Config.PathPrefix}}';
    const sitePrefix = './';
    const sessionIntervalSeconds = 45*60;
    const imgWidth = 400;
    const refreshInterval = 5000;

    async function listBucketContents(bucket, prefix, continuationToken=null, entries=[]) {
      const params = {
        Bucket: bucket,
        ContinuationToken: continuationToken,
        Prefix: prefix
      };

      const data = await s3.listObjectsV2(params).promise();
      entries.push(...data.Contents.map(item => {
        if (prefix) {
          return item.Key.slice(prefix.length);
        }
        return item.Key;
      }));

      if (data.IsTruncated) {
        entries = await listBucketContents(bucket, prefix, data.NextContinuationToken, entries);
      } else {
        entries.sort();
      }

      return entries;
    }

    function timestamp(path) {
      var parts = path.split("/");
      if (parts.length != 5) {
        return null;
      }
      var timeparts = parts[4].split(".")[0].split("-");
      return Date.parse(parts[0] + "-" + parts[1] + "-" + parts[2] + "T" +
                        parts[3] + ":" + timeparts[0] + ":" + timeparts[1] + ".000Z");
    }

    function zeroPad(val, length) {
      var rv = "" + val;
      while (rv.length < length) {
        rv = "0" + rv;
      }
      return rv;
    }

    async function refresh(bucket, prefix, img, interval) {
      var now = new Date();
      var subprefix = now.getUTCFullYear() + "/" +
            zeroPad(now.getUTCMonth() + 1, 2) + "/" +
            zeroPad(now.getUTCDate(), 2) + "/" +
            zeroPad(now.getUTCHours(), 2) + "/";
      var entries = await listBucketContents(bucket, prefix + subprefix);
      if (entries.length == 0) { return; }
      var latest = subprefix + entries[entries.length - 1];
      if (interval.entries[interval.entries.length - 1] == latest) {
        return;
      }
      interval.entries.push(latest);
      interval.end = timestamp(latest);
      img.src = latest;
    }

    async function main() {
      const allEntries = await listBucketContents(bucketName, bucketPrefix);

      var beginning = null;
      var last = null;
      var intervals = [];
      var entries = [];
      allEntries.forEach((item) => {
        var ts = timestamp(item);
        if (!ts) { return; }
        if (!beginning) {
          beginning = ts;
          last = ts;
          entries.push(item);
          return;
        }
        if (ts - sessionIntervalSeconds * 1000 <= last) {
            last = ts;
            entries.push(item);
            return;
        }
        intervals.push({beginning: beginning, end: last, entries: entries});
        beginning = ts;
        last = ts;
        entries = [item];
        return;
      });
      if (beginning) {
        intervals.push({beginning: beginning, end: last, entries: entries})
      }

      function dateOrLive(d) {
        if ((new Date()).getTime() - sessionIntervalSeconds * 1000 <= d) {
            return "now";
        }
        return (new Date(d)).toLocaleString();
      }

      var bucketContentList = document.getElementById("bucket-content");
      intervals.reverse().forEach((interval) => {
        var listItem = document.createElement("li");
        var img = document.createElement("img");
        img.src = sitePrefix + interval.entries[interval.entries.length-1];
        img.width = imgWidth;
        img.onmousemove = function(ev) {
          var entry = Math.round(ev.offsetX * interval.entries.length / imgWidth);
          if (interval.entries[entry]) {
            img.src = sitePrefix + interval.entries[entry];
          }
        };
        var text = document.createElement("p");
        text.textContent = "" + (new Date(interval.beginning)).toLocaleString() + " to " + dateOrLive(interval.end);
        listItem.appendChild(text);
        listItem.appendChild(img);
        bucketContentList.appendChild(listItem);
        if (dateOrLive(interval.end) == "now") {
          function tick() {
            refresh(bucketName, bucketPrefix, img, interval);
          }
          setInterval(tick, refreshInterval);
        }
      });
    }
  </script>
</body>
</html>
`))
