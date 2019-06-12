/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"log"
	"net/http"
)

var UseProxy bool
var Socks5Addr string
var BindAddr string
var ListenPort string
var PayloadPath string

var rootCmd = &cobra.Command{
	Use:   "reverseproxy",
	Short: "A http reverse proxy for byr bt",
	Run: func(cmd *cobra.Command, args []string) {
		if !UseProxy {
			Socks5Addr = ""
		}
		rp, err := NewReverseProxy(BindAddr, ListenPort, Socks5Addr, PayloadPath)
		if err != nil {
			log.Fatal(err)
		}


		if !UseProxy {
			Socks5Addr = "NONE"
		}
		log.Printf("Proxy server running on [%s:%s], using socks5 proxy [%s].\n", BindAddr, ListenPort, Socks5Addr)
		err = http.ListenAndServe(BindAddr + ":" + ListenPort, rp)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&UseProxy, "use_proxy", false, "use socks5 proxy or not")
	rootCmd.Flags().StringVarP(&Socks5Addr, "socks5", "", "localhost:1080", "socks5 proxy address [addr:port]")
	rootCmd.Flags().StringVarP(&ListenPort, "port", "", "9988", "server listen port")
	rootCmd.Flags().StringVarP(&BindAddr, "addr", "", "localhost", "server bind address")
	rootCmd.Flags().StringVarP(&PayloadPath, "payload_path", "", "__payload.js", "js payload path")
}
