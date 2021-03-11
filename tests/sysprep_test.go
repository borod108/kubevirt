/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package tests_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/flags"
)

const (
	windowsSealedDisk = "windows-disk"
	//windowsFirmware    = "5d307ca9-b3ef-428c-8861-06e72d69f223"
	windowsSysprepVMIUser     = "Admin"
	windowsSysprepVMIPassword = "Gauranga"
)

const (
	answerFile = `
    <?xml version="1.0" encoding="utf-8"?>
    <unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="windowsPE">
    <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <SetupUILanguage>
    <UILanguage>en-US</UILanguage>
    </SetupUILanguage>
    <InputLocale>0c09:00000409</InputLocale>
    <SystemLocale>en-US</SystemLocale>
    <UILanguage>en-US</UILanguage>
    <UILanguageFallback>en-US</UILanguageFallback>
    <UserLocale>en-AU</UserLocale>
    </component>
    <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <DiskConfiguration>
    <Disk wcm:action="add">
    <CreatePartitions>
    <CreatePartition wcm:action="add">
    <Order>1</Order>
    <Type>Primary</Type>
    <Size>100</Size>
    </CreatePartition>
    <CreatePartition wcm:action="add">
    <Extend>true</Extend>
    <Order>2</Order>
    <Type>Primary</Type>
    </CreatePartition>
    </CreatePartitions>
    <ModifyPartitions>
    <ModifyPartition wcm:action="add">
    <Active>true</Active>
    <Format>NTFS</Format>
    <Label>System Reserved</Label>
    <Order>1</Order>
    <PartitionID>1</PartitionID>
    <TypeID>0x27</TypeID>
    </ModifyPartition>
    <ModifyPartition wcm:action="add">
    <Active>true</Active>
    <Format>NTFS</Format>
    <Label>OS</Label>
    <Letter>C</Letter>
    <Order>2</Order>
    <PartitionID>2</PartitionID>
    </ModifyPartition>
    </ModifyPartitions>
    <DiskID>0</DiskID>
    <WillWipeDisk>true</WillWipeDisk>
    </Disk>
    </DiskConfiguration>
    <ImageInstall>
    <OSImage>
    <InstallTo>
    <DiskID>0</DiskID>
    <PartitionID>2</PartitionID>
    </InstallTo>
    <InstallToAvailablePartition>false</InstallToAvailablePartition>
    </OSImage>
    </ImageInstall>
    <UserData>
    <AcceptEula>true</AcceptEula>
    <FullName>admin</FullName>
    <Organization></Organization>
    </UserData>
    <EnableFirewall>true</EnableFirewall>
    </component>
    </settings>
    <settings pass="offlineServicing">
    <component name="Microsoft-Windows-LUA-Settings" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <EnableLUA>false</EnableLUA>
    </component>
    </settings>
    <settings pass="generalize">
    <component name="Microsoft-Windows-Security-SPP" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <SkipRearm>1</SkipRearm>
    </component>
    </settings>
    <settings pass="specialize">
    <component name="Microsoft-Windows-International-Core" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <InputLocale>0c09:00000409</InputLocale>
    <SystemLocale>en-AU</SystemLocale>
    <UILanguage>en-AU</UILanguage>
    <UILanguageFallback>en-AU</UILanguageFallback>
    <UserLocale>en-AU</UserLocale>
    </component>
    <component name="Microsoft-Windows-Security-SPP-UX" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <SkipAutoActivation>true</SkipAutoActivation>
    </component>
    <component name="Microsoft-Windows-SQMApi" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <CEIPEnabled>0</CEIPEnabled>
    </component>
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <ComputerName>ANAME-PC</ComputerName>
    <ProductKey>Y4TGP-NPTV9-HTC2H-7MGQ3-DV4TW</ProductKey>
    </component>
    </settings>
    <settings pass="oobeSystem">
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <AutoLogon>
    <Password>
    <Value>Meslo123</Value>
    <PlainText>true</PlainText>
    </Password>
    <Enabled>false</Enabled>
    <Username>admin</Username>
    </AutoLogon>
    <OOBE>
    <HideEULAPage>true</HideEULAPage>
    <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
    <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
    <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
    <NetworkLocation>Work</NetworkLocation>
    <ProtectYourPC>1</ProtectYourPC>
    <SkipUserOOBE>true</SkipUserOOBE>
    <SkipMachineOOBE>true</SkipMachineOOBE>
    </OOBE>
    <UserAccounts>
    <LocalAccounts>
    <LocalAccount wcm:action="add">
    <Password>
    <Value>Meslo123</Value>
    <PlainText>true</PlainText>
    </Password>
    <Description></Description>
    <DisplayName>admin</DisplayName>
    <Group>Administrators</Group>
    <Name>admin</Name>
    </LocalAccount>
    </LocalAccounts>
    </UserAccounts>
    <RegisteredOrganization></RegisteredOrganization>
    <RegisteredOwner>admin</RegisteredOwner>
    <DisableAutoDaylightTimeSet>false</DisableAutoDaylightTimeSet>
    <TimeZone>AUS Eastern Standard Time</TimeZone>
    <VisualEffects>
    <SystemDefaultBackgroundColor>2</SystemDefaultBackgroundColor>
    </VisualEffects>
    </component>
    <component name="Microsoft-Windows-ehome-reg-inf" processorArchitecture="x86" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="NonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <RestartEnabled>true</RestartEnabled>
    </component>
    <component name="Microsoft-Windows-ehome-reg-inf" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="NonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <RestartEnabled>true</RestartEnabled>
    </component>
    </settings>
    </unattend>
    `
)

// const (
//     winrmCli    = "winrmcli"
//     winrmCliCmd = "winrm-cli"
// )

var getWindowsSysprepVMISpec = func() v1.VirtualMachineInstanceSpec {
	gracePeriod := int64(0)
	spinlocks := uint32(8191)
	firmware := types.UID(windowsFirmware)
	_false := false
	return v1.VirtualMachineInstanceSpec{
		TerminationGracePeriodSeconds: &gracePeriod,
		Domain: v1.DomainSpec{
			CPU: &v1.CPU{Cores: 2},
			Features: &v1.Features{
				ACPI: v1.FeatureState{},
				APIC: &v1.FeatureAPIC{},
				Hyperv: &v1.FeatureHyperv{
					Relaxed:   &v1.FeatureState{},
					VAPIC:     &v1.FeatureState{},
					Spinlocks: &v1.FeatureSpinlocks{Retries: &spinlocks},
				},
			},
			Clock: &v1.Clock{
				ClockOffset: v1.ClockOffset{UTC: &v1.ClockOffsetUTC{}},
				Timer: &v1.Timer{
					HPET:   &v1.HPETTimer{Enabled: &_false},
					PIT:    &v1.PITTimer{TickPolicy: v1.PITTickPolicyDelay},
					RTC:    &v1.RTCTimer{TickPolicy: v1.RTCTickPolicyCatchup},
					Hyperv: &v1.HypervTimer{},
				},
			},
			Firmware: &v1.Firmware{UUID: firmware},
			Resources: v1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					k8sv1.ResourceMemory: resource.MustParse("2048Mi"),
				},
			},
			Devices: v1.Devices{
				Disks: []v1.Disk{
					{
						Name:       windowsSealedDisk,
						DiskDevice: v1.DiskDevice{Disk: &v1.DiskTarget{Bus: "sata"}},
					},
					{
						Name:       "sysprep",
						DiskDevice: v1.DiskDevice{CDRom: &v1.CDRomTarget{Bus: "sata"}},
					},
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: windowsSealedDisk,
				VolumeSource: v1.VolumeSource{
					Ephemeral: &v1.EphemeralVolumeSource{
						PersistentVolumeClaim: &k8sv1.PersistentVolumeClaimVolumeSource{
							ClaimName: tests.DiskWindows,
						},
					},
				},
			},
			{
				Name: "sysprep",
				VolumeSource: v1.VolumeSource{
					Sysprep: &v1.SysprepSource{
						ConfigMap: &k8sv1.LocalObjectReference{
							Name: "sysprepautounattend",
						},
					},
				},
			},
		},
	}

}

const (
	windowsSysprepedVMIUser     = "Admin"    // "Administrator" //
	windowsSysprepedVMIPassword = "Meslo123" //      "Heslo123"      //
)

var _ = Describe("[Serial]Windows VirtualMachineInstance", func() {
	var err error
	var virtClient kubecli.KubevirtClient

	var windowsVMI *v1.VirtualMachineInstance

	BeforeEach(func() {
		virtClient, err = kubecli.GetKubevirtClient()
		tests.PanicOnError(err)
		tests.BeforeTestCleanup()
		tests.SkipIfNoWindowsImage(virtClient)
		tests.CreatePVC(tests.OSWindows, "30Gi", tests.Config.StorageClassWindows, true)
		tests.CreateConfigMap("sysprepautounattend", map[string]string{"Autounattend.xml": answerFile, "Unattend.xml": answerFile})
		windowsVMI = tests.NewRandomVMI()
		windowsVMI.Spec = getWindowsSysprepVMISpec()
		tests.AddExplicitPodNetworkInterface(windowsVMI)
		windowsVMI.Spec.Domain.Devices.Interfaces[0].Model = "e1000"
	})

	Context("[ref_id:???]should create the Admin user as specifide in the Autounattend.xml", func() {
		var winrmcliPod *k8sv1.Pod
		var cli []string
		var output string
		var vmiIp string

		BeforeEach(func() {
			By("Creating winrm-cli pod for the future use")
			winrmcliPod = &k8sv1.Pod{
				ObjectMeta: metav1.ObjectMeta{GenerateName: winrmCli},
				Spec: k8sv1.PodSpec{
					Containers: []k8sv1.Container{
						{
							Name:    winrmCli,
							Image:   fmt.Sprintf("%s/%s:%s", flags.KubeVirtUtilityRepoPrefix, winrmCli, flags.KubeVirtUtilityVersionTag),
							Command: []string{"sleep"},
							Args:    []string{"3600"},
						},
					},
				},
			}
			winrmcliPod, err = virtClient.CoreV1().Pods(tests.NamespaceTestDefault).Create(context.Background(), winrmcliPod, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			By("Starting the windows VirtualMachineInstance")
			windowsVMI, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(windowsVMI)
			Expect(err).ToNot(HaveOccurred())
			tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)

			windowsVMI, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(windowsVMI.Name, &metav1.GetOptions{})
			vmiIp = windowsVMI.Status.Interfaces[0].IP
			cli = []string{
				winrmCliCmd,
				"-hostname",
				vmiIp,
				"-username",
				windowsSysprepedVMIUser,
				"-password",
				windowsSysprepedVMIPassword,
			}
		})

		FIt("[test_id:???1]should have worked", func() {
			command := append(cli, "echo works")
			By(fmt.Sprintf("Running \"%s\" command via winrm-cli", command))
			fmt.Println("hereoo!!!!!!1")
			Eventually(func() error {
				output, err = tests.ExecuteCommandOnPod(
					virtClient,
					winrmcliPod,
					winrmcliPod.Spec.Containers[0].Name,
					command,
				)
				return err
			}, time.Minute*5, time.Second*15).ShouldNot(HaveOccurred())
			By("Checking that the Windows VirtualMachineInstance has expected UUID")
			fmt.Println("hereoo!!!!!!2")
			fmt.Println(output)
			Expect(output).Should(ContainSubstring("works"))
		}, 720)
	})
})
