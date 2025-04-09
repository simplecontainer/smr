package dns

import (
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/domains"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func (r *Records) ListenRecords() {
	for data := range r.Records {
		d := Distributed{}
		err := json.Unmarshal(data.Val, &d)

		format := f.NewFromString(data.Key)
		acks.ACKS.Ack(format.GetUUID())

		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}

		switch d.Action {
		case AddRecord:
			r.processRecord(d.Domain, d.IP, r.AddAndSave)
		case RemoveRecord:
			r.processRecord(d.Domain, d.IP, r.RemoveAndSave)
		}
	}
}

func (r *Records) processRecord(domain, ip string, actionFunc func(string, string)) {
	actionFunc(domain, ip)
}

func (r *Records) Propose(domain string, ip string, action uint8) error {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_DNS, "dns", "internal", domain)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)

	d := domains.NewFromString(domain)

	bytes, err := json.Marshal(Distributed{
		Domain:   d.ToString(),
		Headless: d.ToHeadles(),
		IP:       ip,
		Action:   action,
	})

	if err != nil {
		return err
	}

	return obj.Wait(format, bytes)
}

func (r *Records) AddAndSave(domain string, ip string) {
	r.saveRecord(r.AddARecord, domain, ip)
}

func (r *Records) RemoveAndSave(domain string, ip string) {
	r.saveRecord(r.RemoveARecord, domain, ip)
}

func (r *Records) saveRecord(actionFunc func(string, string) ([]byte, error), domain string, ip string) {
	addresses, err := actionFunc(domain, ip)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if addresses == nil {
		_, err = r.Remove(addresses, domain)
	} else {
		err = r.Save(addresses, domain)
	}

	if err != nil {
		logger.Log.Error(err.Error())
	}
}
