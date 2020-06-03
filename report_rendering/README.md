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
![dsnet report table](https://user-images.githubusercontent.com/1956773/83570601-439a2980-a51e-11ea-874d-fea32f05abb4.png)

