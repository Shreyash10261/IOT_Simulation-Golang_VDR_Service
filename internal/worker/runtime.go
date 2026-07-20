package worker

import (
	"fmt"
	"log"
	"os/exec"
)

// AssignIPToBridge attempts to run the command: `ip addr add <ip>/24 dev <interfaceName>`
// to register the IP on the bridge interface. It fails gracefully with a warning
// log if it lacks elevated permissions or the interface does not exist.
func AssignIPToBridge(ip string, interfaceName string) error {
	cmd := exec.Command("ip", "addr", "add", ip+"/24", "dev", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip command failed: %v (output: %q)", err, string(output))
	}
	log.Printf("Successfully bound IP %s to interface %s via shell.", ip, interfaceName)
	return nil
}
