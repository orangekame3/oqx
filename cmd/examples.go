package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

const bellProgram = `OPENQASM 3; qubit[2] q; bit[2] c; h q[0]; cnot q[0], q[1]; c = measure q;`

var exampleDevice string
var exampleShots int

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Print request body examples",
}

var examplesSubmitJobCmd = &cobra.Command{
	Use:   "submit-job",
	Short: "Print a jobs submit request body",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := samplingJobBody(exampleDevice, bellProgram, "Bell State Sampling", "Bell State Sampling Example", exampleShots)
		if err != nil {
			return err
		}
		var value any
		if err := json.Unmarshal(body, &value); err != nil {
			return err
		}
		return printValue(cmd, value)
	},
}

func init() {
	examplesSubmitJobCmd.Flags().StringVar(&exampleDevice, "device", "qulacs", "Device ID")
	examplesSubmitJobCmd.Flags().IntVar(&exampleShots, "shots", 1000, "Number of shots")
	examplesCmd.AddCommand(examplesSubmitJobCmd)
}
