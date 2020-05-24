<?php

// Thanks to github.com/Write. See https://github.com/naggie/dsnet/issues/4#issuecomment-632928158 for background.

/* Look for dsnetreport.json in current directory */
/* Change "ReportFile": "/var/lib/dsnetreport.json" accordingly */
/* Also add a crontab to run "dsnet report" to refresh the dsnetreport file */
$json = file_get_contents(__DIR__.'/dsnetreport.json');

$decoded_json = json_decode($json, true);
$ip = $decoded_json['ExternalIP'];
$total = $decoded_json['PeersTotal'];
$online = $decoded_json['PeersOnline'];
$peers = $decoded_json['Peers'];

?>
<!DOCTYPE html>
<html>
<head>
	<link rel="stylesheet" href="https://newcss.net/new.min.css">
	<style>
		html {
			-webkit-text-size-adjust: none;
			touch-action: none;
		}
		body {
			max-width: 800px;
		}
		table {
			white-space: nowrap;
			display: block;
			overflow-x: auto;
		}
		small {
			font-size: 0.75rem;
		}
		mark {
			color: whitesmoke;
		}
		muted {
			color: #0000008c;
		}
		.square {
		  height: 15px;
		  width: 15px;
		  margin: auto;
		}
		.on {
		  background-color: #22e665;
		  box-shadow: 0px 0px 10px #22e665a8;
		}
		.off {
		  background-color: #e62259;
		}
		tr {
		    background: var(--nc-bg-1);
		}
		@media (prefers-color-scheme: dark) {
			:root {
				--nc-tx-1: #d7e6fb;
				--nc-tx-2: #c7d4e6;
			}
			mark {
				color: var(--nc-ac-tx);
			}
			muted {
				color: #ffffff75;
			}
		}
	</style>
	<meta name="viewport" content="width=device-width, initial-scale=1.0, minimium-scale=1.0, user-scalable=no" />
	<title>WireGuard Peers</title>
</head>
<body>
	<header>
		<h2>
			WireGuard
		</h2>
	</header>
	<p>WireGuard Peers</p>
	<table>
        <thead>
            <tr>
                <th>Hostname</th>
                <th>Status</th>
                <th>IP</th>
                <th>Owner</th>
                <th>Description</th>
                <th>Up</th>
                <th>Down</th>
            </tr>
        </thead>
        <tbody>
        	<?php
			foreach($peers as $key=>$value){
			?>
			<tr>
				<?php
				echo '<td>'.$value['Hostname'].'</td>';
				echo '<td>'.($value['Online'] ? '<div class="square on"></div>' : '<div class="square off"></div>').'</td>';
				echo '<td>'.$value['IP'].'</td>';
				echo '<td>'.$value['Owner'].'</td>';
				echo '<td>'.$value['Description'].'</td>';
				echo '<td>'.$value['ReceiveBytesSI'].'</td>';
				echo '<td>'.$value['TransmitBytesSI'].'</td>';
				?>
			</tr>
			<?php
			}
        	?>
        </tbody>
	</table>
	<em>
    		<?php echo $online. ' of '.$total.' devices connected'; ?>
	</em>
	<hr>
	<em>
		<?php
		foreach($peers as $key=>$value) {
			$date_str = substr($value['LastHandshakeTime'], 0, 19);
			$d2 = new DateTime($date_str);
			echo '<muted><small>'.$value['Hostname'].' • Handshake : '. $d2->format("d-m \a\\t H:i:s").'</small></muted><br/>';
		}
		?>
	</em>
</body>
</html>
