package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/controlplaneio/kubectl-kubesec/v2/pkg/kubesec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var statefulsetCmd = &cobra.Command{
	Use:     `statefulset [name]`,
	Short:   "Scans statefulset object",
	Example: `  statefulset podinfo --namespace=default`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("deployment name is required")
		}
		name := args[0]

		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)

		fmt.Println("scanning statefulset", name, "in namespace", namespace)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		ss, err := kubeClient.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ss.TypeMeta = metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		}
		err = serializer.Encode(ss, writer)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := writer.Flush(); err != nil {
			return err
		}

		kc := kubesec.NewClient(scanURL, scanTimeOut)
		rs, err := kc.ScanDefinition(buffer)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := rs.Dump(os.Stdout); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return nil
	},
}
