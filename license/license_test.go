package license

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const LicenseB32 = "FT7YCAYBAEDUY2LDMVXHGZIB76BAAAIDAECEIYLUMEAQUAABAFJAD74EAAAQCUYB76CAAAAABL7YGBIBAL7YMAAAAD7AF7H7QIA74AUPPMRGK3LBNFWCEORCORSXG5CAMV4GC3LQNRSS4Y3PNURCYITFNZSCEORCGIYDCOJNGA2C2MBXKQYTQORRGY5DENZOGU2TSNRYGAZDGNBLGAZDUMBQEIWCE3DJNVUXI4ZCHJNXWITVNZWGS3LJORSWIIR2MZQWY43FFQRG2YLYL5RW65LOOQRDUMRMEJUW45DFOJ3GC3BCHIYTAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORRGAWCE2LOORSXE5TBNQRDUNRQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5FR5SE5LONRUW22LUMVSCEOTGMFWHGZJMEJWWC6C7MNXXK3TUEI5DEMBQGAWCE2LOORSXE5TBNQRDUMZWGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQFQRGS3TUMVZHMYLMEI5DQNRUGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQGAWCE2LOORSXE5TBNQRDUMRVHEZDAMBQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5LUWCEQ3SPFYHI32LMV4SEOS3GAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDAXJMEJJEYIR2NZ2WY3D5AEYQFVEO4FTR5GJ6XT6YL4EU4OOPGP73D6AAH5DKOPIFYPXLA6DNQFHULFQME5SLIEP4KRZYR6KUD2PILIATCAXSNKKQPNJ6O2UUTS7IODUZ6DSXQWZU33UDHIK7LMZ45IMOOKAFQJXJ6MF74RVHNPCZVUFRYOXFZCAAA==="

var licData = LicenseData{
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2, Interval: time.Second},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 10, Interval: time.Minute},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000, Interval: time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 200000, Interval: 24 * time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000000, Interval: 720 * time.Hour},
	},
}

func TestFlow(t *testing.T) {
	// test initialize
	licenseBytes := []byte(LicenseB32)
	err := Initialize(licenseBytes)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(deep.Equal(licData.Limits, LD.Limits)))
	assert.Equal(t, 2, LD.Limits[0].MaxCount)
	assert.Equal(t, 0, LD.Limits[0].LimitState.CurCount)

	// test load
	err = Load()
	assert.NoError(t, err)

	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 1, LD.Limits[0].CurCount)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	time.Sleep(1 * time.Second)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 1, LD.Limits[0].CurCount)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)
	assert.False(t, LD.RL.Allow())
	assert.True(t, LD.RL.Left() > 0)

	// test save
	err = LD.SaveLimitState()
	assert.NoError(t, err)

	LD = LicenseData{}
	err = Load()
	assert.NoError(t, err)
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)

}