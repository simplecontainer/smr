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
	for {
		select {
		case data := <-r.Records:
			// Be careful to not break or return we want this to run forever
			d := Distributed{}
			err := json.Unmarshal(data.Val, &d)

			format := f.NewFromString(data.Key)
			acks.ACKS.Ack(format.GetUUID())

			if err != nil {
				logger.Log.Error(err.Error())
				break
			}

			switch d.Action {
			case AddRecord:
				r.AddAndSave(d.Domain, d.IP)
				r.AddAndSave(d.Headless, d.IP)

				break
			case RemoveRecord:
				r.RemoveAndSave(d.Domain, d.IP)
				r.RemoveAndSave(d.Headless, d.IP)

				break
			}
			break
		}
	}
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
	var addresses []byte
	var err error

	addresses, err = r.AddARecord(domain, ip)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	err = r.Save(addresses, domain)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}
func (r *Records) RemoveAndSave(domain string, ip string) {
	var addresses []byte
	var err error

	addresses, err = r.RemoveARecord(domain, ip)

	if err != nil {
		logger.Log.Error(err.Error())
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
