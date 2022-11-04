This directory contains scripts and templates that can be used to render
`/var/lib/dsnetreport.json`. They are useful for integrating a peer overview
into an existing website or web application.

Most are contributions from other users. If you have a useful addition, please
do a PR.

Most look something like this:

## Hugo shortcode template

* `hugo/dsnetreport.html`: A hugo shortcode for rendering a report.

![dsnet report table](https://raw.githubusercontent.com/naggie/dsnet/master/etc/report.png)

# PHP template
See https://github.com/naggie/dsnet/issues/4#issuecomment-632928158 for background. Courtesy of [@Write](https://github.com/Write)

* `php/dsnetreport.php`: A php file to render a report.

![dsnet report table](https://user-images.githubusercontent.com/541722/82712747-0cf42180-9c89-11ea-92fa-0974a34c5c79.jpg)
![dsnet report table](https://user-images.githubusercontent.com/541722/82712745-0a91c780-9c89-11ea-91a8-828e0be38951.jpg)

# Clientside JavaScript

Courtesy of [@frillip](https://github.com/frillip/)

* `js/dsnetreport.html`: Basic HTML with a `div` to place the table in.
* `js/dsnetreport.js`: Fetches `dsnetreport.json` and renders table.
* `js/dsnetreport.css`: CSS to render the table as per screenshot.

![dsnet report table](https://raw.githubusercontent.com/naggie/dsnet/master/etc/dsnet-report-js.png)

On the command line, you can use [jtbl](https://github.com/kellyjonbrazil/jtbl) (and [jq](https://stedolan.github.io/jq/)) for a nice table rendering with this snippet:

```bash
sudo dsnet report | jq '.Peers' | jtbl
```

The output looks like:
```
╒═════════╤═══════╤══════════╤══════════╤══════════╤══════════╤═════════╤════════╤══════════╤══════════╤══════════╤══════════╤══════════╤══════════╤══════════╕
│ Owner   │ IP6   │ Hostna   │ Descri   │ Online   │ Dorman   │ Added   │ IP     │ Extern   │ Networ   │ LastHa   │ Receiv   │ Transm   │ Receiv   │ Transm   │
│         │       │ me       │ ption    │          │ t        │         │        │ alIP     │ ks       │ ndshak   │ eBytes   │ itByte   │ eBytes   │ itByte   │
│         │       │          │          │          │          │         │        │          │          │ eTime    │          │ s        │ SI       │ sSI      │
╞═════════╪═══════╪══════════╪══════════╪══════════╪══════════╪═════════╪════════╪══════════╪══════════╪══════════╪══════════╪══════════╪══════════╪══════════╡
│ xyz     │       │ eaetl    │ eaetl.   │ True     │ False    │ 2222-0  │ 99.99. │ dddd:d   │ []       │ 1111-1   │ 175995   │ 447007   │ 175.9    │ 32.7 M   │
│         │       │          │ fooo     │          │          │ 2-22T1  │ 99.9   │ dd:ddd   │          │ 1-11T1   │ 424      │ 28       │ MB       │ B        │
│         │       │          │          │          │          │ 2:22:5  │        │ d:dddd   │          │ 1:11:1   │          │          │          │          │
│         │       │          │          │          │          │ 2.2274  │        │ :dddd:   │          │ 1.1111   │          │          │          │          │
│         │       │          │          │          │          │ 22222-  │        │ dddd:d   │          │ 11111-   │          │          │          │          │
│         │       │          │          │          │          │ 22:20   │        │ ddd:dd   │          │ 11:11    │          │          │          │          │
│         │       │          │          │          │          │         │        │ dd       │          │          │          │          │          │          │
├─────────┼───────┼──────────┼──────────┼──────────┼──────────┼─────────┼────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ xyz     │       │ ammedu   │ ammedu   │ True     │ False    │ 2222-0  │ 88.88. │ eeee:e   │ []       │ 1111-1   │ 751670   │ 759741   │ 6.7 GB   │ 727.7    │
│         │       │          │ .mymy.   │          │          │ 2-22T1  │ 88.8   │ eee:ee   │          │ 1-11T1   │ 2852     │ 076      │          │ MB       │
│         │       │          │ com      │          │          │ 2:22:4  │        │ ee:eee   │          │ 1:11:1   │          │          │          │          │
│         │       │          │          │          │          │ 2.2292  │        │ e::e     │          │ 1.1111   │          │          │          │          │
│         │       │          │          │          │          │ 22226-  │        │          │          │ 11111-   │          │          │          │          │
│         │       │          │          │          │          │ 22:20   │        │          │          │ 11:11    │          │          │          │          │
├─────────┼───────┼──────────┼──────────┼──────────┼──────────┼─────────┼────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
...
```