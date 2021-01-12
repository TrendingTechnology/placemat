package virtualbmc

import (
	"context"
	"fmt"
	"net"

	"github.com/cybozu-go/well"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Virtual BMC", func() {
	It("should turn on and off Machine power via IPMI v2.0", func() {
		serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", 9623))
		Expect(err).NotTo(HaveOccurred())
		conn, err := net.ListenUDP("udp", serverAddr)
		Expect(err).NotTo(HaveOccurred())
		well.Go(func(ctx context.Context) error {
			return StartIPMIServer(ctx, conn, &MachineMock{running: false})
		})

		Eventually(func() error {
			ipmipower := well.CommandContext(context.Background(),
				"ipmipower", "--stat", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err := ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: off\n" {
				return fmt.Errorf("ipmipowert stat reponse is not 127.0.0.1: off, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--on", "--wait-until-on", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: ok\n" {
				return fmt.Errorf("ipmipowert on reponse is not 127.0.0.1: ok, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--stat", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: on\n" {
				return fmt.Errorf("ipmipowert stat reponse is not 127.0.0.1: on, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--off", "--wait-until-off", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: ok\n" {
				return fmt.Errorf("ipmipowert off reponse is not 127.0.0.1: ok, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--stat", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: off\n" {
				return fmt.Errorf("ipmipowert stat reponse is not 127.0.0.1: off, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--reset", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: ok\n" {
				return fmt.Errorf("ipmipowert reset reponse is not 127.0.0.1: ok, actual is: %s", string(output))
			}

			ipmipower = well.CommandContext(context.Background(),
				"ipmipower", "--stat", "-u", "cybozu", "-p", "cybozu", "-h", "127.0.0.1:9623", "-D", "LAN_2_0")
			output, err = ipmipower.Output()
			if err != nil {
				return err
			}
			if string(output) != "127.0.0.1: on\n" {
				return fmt.Errorf("ipmipowert stat reponse is not 127.0.0.1: on, actual is: %s", string(output))
			}

			return err
		}).Should(Succeed())

		well.Stop()
	})
})

type MachineMock struct {
	running bool
}

func (v *MachineMock) IsRunning() bool {
	return v.running
}

func (v *MachineMock) PowerOn() error {
	v.running = true
	return nil
}

func (v *MachineMock) PowerOff() error {
	v.running = false
	return nil
}