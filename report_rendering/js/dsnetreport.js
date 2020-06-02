// Simple javascript to build a HTML table from 'dsnetreport.json'

// Declare our headings
var header_list = ["Hostname", "Status", "IP", "Owner", "Description", "Up", "Down"];

function build_table() {
  // Get our div
  var report = document.getElementById("dsnetreport");
  report.innerHTML = "";
  // Make our table
  var table = document.createElement("table");
  var header = table.createTHead();
  var row = header.insertRow();
  header_list.forEach(function(heading, index) {
    var cell = row.insertCell();
    // By default, insertCell() creates elements as '<td>' even if in a <thead> for no reason
    cell.outerHTML = "<th>" + heading + "</th>";
  });
  // Create a summary to go at the bottom
  var devices_online = document.createElement("em")

  // By default, this looks for dsnetreport.json in the current directory
  fetch("dsnetreport.json")
    .then(response => response.json())
    .then(data => {
      // Create our summary statement
      devices_online.innerHTML = data.PeersOnline + " of " + data.PeersTotal + " devices connected"
      // Iterate over the peers
      data.Peers.forEach(function(peer, index) {
        // Create the row
        var row = table.insertRow();
        row.id = "peer-" + peer.Hostname;
        row.classList.add("peer")
        // Different colour text if the peer is dormant
        if (peer.Dormant) {
          row.classList.add("dormant")
        }

        // Hostname
        var hostname = row.insertCell();
        hostname.classList.add("hostname")
        hostname.innerHTML = peer.Hostname;
        hostname.title = peer.Hostname + "." + data.Domain;

        // Status
        var status = row.insertCell();
        status.classList.add("status")
        status.setAttribute("nowrap", true)
        // Set indicators based on online status
        if (peer.Online) {
          status.title = "Handshake in last 3 minutes";
          status.classList.add("indicator-green")
          status.innerHTML = "online";
        } else {
          handshake = new Date(peer.LastHandshakeTime);
          // Add some information about when the peer was last seen
          status.title = "No handshake since since " + handshake.toLocaleString();
          status.classList.add("indicator-null")
          status.innerHTML = "offline";
        }

        // IP
        // Could also have external IP as a title?
        var IP = row.insertCell();
        IP.classList.add("ip")
        IP.innerHTML = peer.IP;

        // Owner
        var owner = row.insertCell();
        owner.classList.add("owner")
        owner.innerHTML = peer.Owner;

        // Description
        var desc = row.insertCell();
        desc.classList.add("description")
        desc.innerHTML = peer.Description;

        // Data up in SI units
        var data_up = row.insertCell();
        data_up.classList.add("up")
        data_up.innerHTML = peer.ReceiveBytesSI;

        // Data down in SI units
        var data_down = row.insertCell();
        data_down.classList.add("down")
        data_down.innerHTML = peer.TransmitBytesSI;

      });
    }).catch(error => {
      // If we encounter an error, don't do anything useful, just complain
      console.log(error);
    });
  // Add the table to the div
  report.appendChild(table);
  // Add the summary to the div
  report.appendChild(devices_online);
}

// Build the table when the page has loaded
document.addEventListener("DOMContentLoaded", function() {
  build_table();
}, false);
