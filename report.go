package dsnet

import (
	"golang.zx2c4.com/wireguard/wgctrl"
)

func Report() {
	conf := MustLoadDsnetConfig()

	wg, err := wgctrl.New()
	check(err)
	defer wg.Close()

	dev, err := wg.Device(conf.InterfaceName)

	if err != nil {
		ExitFail("Could not retrieve device '%s' (%v)", conf.InterfaceName, err)
	}

	oldReport := MustLoadDsnetReport()
	report := GenerateReport(dev, conf, oldReport)
	report.MustSave(conf.ReportFile)
}
