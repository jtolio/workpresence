# workpresence

one thing i miss about being in the office (because i'm a procrastinator and have bad executive function)
is the feeling of other people in the office occasionally looking over my shoulder. it's a good motivator
to stay on task. i want to stay on task, but facebook has psychologists trying to get me to go refresh a
feed (my last solution to this problem is https://github.com/jtolio/twitoderm)

so this is my new thing. once you have the storj uplink CLI set up, you make
a bucket:

```
uplink mb sj://bucket
```

then you generate listing credentials:

```
uplink share --register --disallow-reads --disallow-writes --disallow-deletes sj://bucket/path-prefix/
```

you run workpresence like:

```
go install github.com/jtolio/workpresence@latest
workpresence \
  --screenshots.command "command-that-makes-a-png-screenshot-with-filename {{.Output}}" \
  --storage.uplink-access "your-general-access-grant" \
  --storage.bucket "bucket" \
  --storage.path-prefix "path-prefix/" \
  --storage.listing-access-key "listing-access-key" \
  --storage.listing-secret-key "listing-secret-key"
```

and then to generate DNS records:

```
uplink share --dns=workpres.yourdomain.com --tls sj://bucket/path-prefix/
```

and https://workpres.yourdomain.com will host a livestream of snapshots of your desktop. 

There is a control page at http://localhost:3333/ where you can pause/resume the snapshot generation
if you need to do something private for a second.

### License

Released under Apache v2 license. See LICENSE file.
