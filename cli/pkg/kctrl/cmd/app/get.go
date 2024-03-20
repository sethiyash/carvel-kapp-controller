// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"github.com/cppforlife/color"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd/core"
	"github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/logger"
	kcv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	NamespaceFlags cmdcore.NamespaceFlags
	Name           string

	columns *[]string
}

func NewGetOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger, columns *[]string) *GetOptions {
	return &GetOptions{ui: ui, depsFactory: depsFactory, logger: logger, columns: columns}
}

func NewGetCmd(o *GetOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "Get details for app",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppManagementCommandsHelpGroup.Key: cmdcore.AppManagementCommandsHelpGroup.Value,
		},
	}

	o.NamespaceFlags.Set(cmd, flagsFactory)
	cmd.Flags().StringVarP(&o.Name, "app", "a", "", "Set app name (required)")

	return cmd
}

func (o *GetOptions) Run() error {
	if len(o.Name) == 0 {
		return fmt.Errorf("Expected app name to be non empty")
	}

	client, err := o.depsFactory.KappCtrlClient()
	if err != nil {
		return err
	}

	app, err := client.KappctrlV1alpha1().Apps(o.NamespaceFlags.Name).Get(context.Background(), o.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	status, isFailing := appStatus(app)

	failingStageHeader := uitable.NewHeader("Failing Stage")
	failingStageHeader.Hidden = len(app.Status.UsefulErrorMessage) == 0
	errorMessageHeader := uitable.NewHeader("Useful Error Message")
	errorMessageHeader.Hidden = len(app.Status.UsefulErrorMessage) == 0

	table := uitable.Table{
		Transpose: true,

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Service Account"),
			uitable.NewHeader("Status"),
			uitable.NewHeader("Owner References"),
			uitable.NewHeader("Conditions"),
			failingStageHeader,
			errorMessageHeader,
		},

		Rows: [][]uitable.Value{{
			uitable.NewValueString(app.Namespace),
			uitable.NewValueString(app.Name),
			uitable.NewValueString(app.Spec.ServiceAccountName),
			uitable.ValueFmt{V: uitable.NewValueString(status), Error: isFailing},
			uitable.NewValueInterface(o.formatOwnerReferences(app.OwnerReferences)),
			uitable.NewValueInterface(app.Status.Conditions),
			uitable.NewValueString(o.failingStage(app.Status)),
			uitable.NewValueString(color.RedString(app.Status.UsefulErrorMessage)),
		}},
	}

	return cmdcore.PrintTable(o.ui, table, o.columns)
}

func (o *GetOptions) formatOwnerReferences(references []metav1.OwnerReference) []string {
	var referenceList []string

	for _, reference := range references {
		referenceList = append(referenceList, fmt.Sprintf("%s/%s/%s", reference.APIVersion, reference.Kind, reference.Name))
	}

	return referenceList
}

// TODO: Do we need to check observed generation here?
func (o *GetOptions) failingStage(status kcv1alpha1.AppStatus) string {
	if status.Fetch.ExitCode != 0 {
		return "fetch"
	}
	if status.Template.ExitCode != 0 {
		return "template"
	}
	if status.Deploy.ExitCode != 0 {
		return "deploy"
	}
	return ""
}
