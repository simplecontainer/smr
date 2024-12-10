package dns

import "errors"

func NewARecord() *ARecord {
	return &ARecord{
		[]string{},
	}
}

func (AR *ARecord) Append(ip string) {
	for _, i := range AR.IPs {
		if i == ip {
			return
		}
	}

	AR.IPs = append(AR.IPs, ip)
}

func (AR *ARecord) Remove(ip string) error {
	for i, ARip := range AR.IPs {
		if ARip == ip {
			AR.IPs = append(AR.IPs[:i], AR.IPs[i+1:]...)
			return nil
		}
	}

	return errors.New("remove failed: not found A record")
}
