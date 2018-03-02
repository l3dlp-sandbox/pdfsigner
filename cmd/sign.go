package cmd

import (
	"bitbucket.org/digitorus/pdfsigner/files"
	"github.com/spf13/cobra"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign command",
	Long:  `Long multiline description here`,
}

var signPEMCmd = &cobra.Command{
	Use:   "pem",
	Short: "Sign PDF with PEM formatted certificate",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		c := signerConfig{}
		bindSignerFlagsToConfig(cmd, &c)
		c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

var signPKSC11Cmd = &cobra.Command{
	Use:   "pksc11",
	Short: "Signs PDF with PSKC11",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		c := signerConfig{}
		bindSignerFlagsToConfig(cmd, &c)
		c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

var signBySignerNameCmd = &cobra.Command{
	Use:   "signer",
	Short: "Signs PDF with signer from the config",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, filePatterns []string) {
		requireConfig(cmd)

		c := getSignerConfigByName(signerNameFlag)
		bindSignerFlagsToConfig(cmd, &c)

		switch c.Type {
		case "pem":
			c.SignData.SetPEM(c.CrtPath, c.KeyPath, c.CrtChainPath)
		case "pksc11":
			c.SignData.SetPKSC11(c.LibPath, c.Pass, c.CrtChainPath)
		}

		files.SignFilesByPatterns(filePatterns, outputPathFlag, c.SignData)
	},
}

func init() {
	RootCmd.AddCommand(signCmd)

	//PEM sign command
	signCmd.AddCommand(signPEMCmd)
	parseCommonFlags(signPEMCmd)
	parseOutputPathFlag(signPEMCmd)
	parsePEMCertificateFlags(signPEMCmd)

	//PKSC11 sign command
	signCmd.AddCommand(signPKSC11Cmd)
	parseCommonFlags(signPKSC11Cmd)
	parseOutputPathFlag(signPKSC11Cmd)
	parsePKSC11CertificateFlags(signPKSC11Cmd)

	// sign with signer from config inputFile
	signCmd.AddCommand(signBySignerNameCmd)
	parseSignerName(signBySignerNameCmd)
	parseOutputPathFlag(signBySignerNameCmd)
	parsePEMCertificateFlags(signBySignerNameCmd)
	parsePKSC11CertificateFlags(signBySignerNameCmd)
}