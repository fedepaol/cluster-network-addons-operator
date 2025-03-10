package test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NetworkAddonsConfig", func() {
	Context("when there is no pre-existing Config", func() {
		DescribeTable("should succeed deploying single component",
			func(configSpec opv1alpha1.NetworkAddonsConfigSpec, components []Component) {
				testConfigCreate(configSpec, components)

				// Make sure that deployed components remain up and running
				CheckConfigCondition(ConditionAvailable, ConditionTrue, CheckImmediately, time.Minute)
			},
			Entry(
				"Empty config",
				opv1alpha1.NetworkAddonsConfigSpec{},
				[]Component{},
			),
			Entry(
				LinuxBridgeComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					LinuxBridge: &opv1alpha1.LinuxBridge{},
				},
				[]Component{LinuxBridgeComponent},
			),
			Entry(
				MultusComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					Multus: &opv1alpha1.Multus{},
				},
				[]Component{MultusComponent},
			),
			Entry(
				SriovComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					Sriov:  &opv1alpha1.Sriov{},
					Multus: &opv1alpha1.Multus{},
				},
				[]Component{SriovComponent, MultusComponent},
			),
			Entry(
				NMStateComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					NMState: &opv1alpha1.NMState{},
				},
				[]Component{NMStateComponent},
			),
			Entry(
				KubeMacPoolComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					KubeMacPool: &opv1alpha1.KubeMacPool{},
				},
				[]Component{KubeMacPoolComponent},
			),
			Entry(
				OvsComponent.ComponentName,
				opv1alpha1.NetworkAddonsConfigSpec{
					Ovs: &opv1alpha1.Ovs{},
				},
				[]Component{OvsComponent},
			),
		)

		It("should be able to deploy all components at once", func() {
			components := []Component{
				MultusComponent,
				LinuxBridgeComponent,
				KubeMacPoolComponent,
				SriovComponent,
				NMStateComponent,
				OvsComponent,
			}
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{
				KubeMacPool: &opv1alpha1.KubeMacPool{},
				LinuxBridge: &opv1alpha1.LinuxBridge{},
				Multus:      &opv1alpha1.Multus{},
				Sriov:       &opv1alpha1.Sriov{},
				NMState:     &opv1alpha1.NMState{},
				Ovs:         &opv1alpha1.Ovs{},
			}
			testConfigCreate(configSpec, components)
		})

		It("should be able to deploy all components one by one", func() {
			configSpec := opv1alpha1.NetworkAddonsConfigSpec{}
			components := []Component{}

			// Deploy initial empty config
			testConfigCreate(configSpec, components)

			// Deploy Multus component
			configSpec.Multus = &opv1alpha1.Multus{}
			components = append(components, MultusComponent)
			testConfigUpdate(configSpec, components)

			// Add Linux bridge component
			configSpec.LinuxBridge = &opv1alpha1.LinuxBridge{}
			components = append(components, LinuxBridgeComponent)
			testConfigUpdate(configSpec, components)

			// Add KubeMacPool component
			configSpec.KubeMacPool = &opv1alpha1.KubeMacPool{}
			components = append(components, KubeMacPoolComponent)
			testConfigUpdate(configSpec, components)

			// Add SR-IOV component
			configSpec.Sriov = &opv1alpha1.Sriov{}
			components = append(components, SriovComponent)
			testConfigUpdate(configSpec, components)

			// Add NMState component
			configSpec.NMState = &opv1alpha1.NMState{}
			components = append(components, NMStateComponent)
			testConfigUpdate(configSpec, components)

			// Add Ovs component
			configSpec.Ovs = &opv1alpha1.Ovs{}
			components = append(components, OvsComponent)
			testConfigUpdate(configSpec, components)
		})
	})

	Context("when all components are already deployed", func() {
		components := []Component{
			MultusComponent,
			LinuxBridgeComponent,
			SriovComponent,
			NMStateComponent,
			KubeMacPoolComponent,
			OvsComponent,
		}
		configSpec := opv1alpha1.NetworkAddonsConfigSpec{
			LinuxBridge: &opv1alpha1.LinuxBridge{},
			Multus:      &opv1alpha1.Multus{},
			Sriov:       &opv1alpha1.Sriov{},
			NMState:     &opv1alpha1.NMState{},
			KubeMacPool: &opv1alpha1.KubeMacPool{},
			Ovs:         &opv1alpha1.Ovs{},
		}

		BeforeEach(func() {
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		It("should remain in Available condition after applying the same config", func() {
			UpdateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, CheckImmediately, 30*time.Second)
		})

		It("should be able to remove all of them by removing the config", func() {
			DeleteConfig()
			CheckComponentsRemoval(components)
		})

		It("should be able to remove the config and create it again", func() {
			DeleteConfig()
			CreateConfig(configSpec)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, 30*time.Second)
		})
	})

	Context("when kubeMacPool is deployed", func() {
		BeforeEach(func() {
			By("Deploying KubeMacPool")
			config := opv1alpha1.NetworkAddonsConfigSpec{KubeMacPool: &opv1alpha1.KubeMacPool{}}
			CreateConfig(config)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		})

		It("should modify the MAC range after being redeployed ", func() {
			oldRangeStart, oldRangeEnd := CheckUnicastAndValidity()
			By("Redeploying KubeMacPool")
			DeleteConfig()
			CheckComponentsRemoval(AllComponents)

			config := opv1alpha1.NetworkAddonsConfigSpec{KubeMacPool: &opv1alpha1.KubeMacPool{}}
			CreateConfig(config)
			CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			rangeStart, rangeEnd := CheckUnicastAndValidity()

			By("Comparing the ranges")
			Expect(rangeStart).ToNot(Equal(oldRangeStart))
			Expect(rangeEnd).ToNot(Equal(oldRangeEnd))
		})
	})
})

func testConfigCreate(configSpec opv1alpha1.NetworkAddonsConfigSpec, components []Component) {
	CreateConfig(configSpec)
	checkConfigChange(components)
}

func testConfigUpdate(configSpec opv1alpha1.NetworkAddonsConfigSpec, components []Component) {
	UpdateConfig(configSpec)
	checkConfigChange(components)
}

func checkConfigChange(components []Component) {
	if len(components) > 0 {
		// If there are any components to deploy wait until Progressing condition is reported
		CheckConfigCondition(ConditionProgressing, ConditionTrue, time.Minute, CheckDoNotRepeat)

		// Wait until Available condition is reported. It may take a few minutes the first time
		// we are pulling component images to the Node
		CheckConfigCondition(ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
		CheckConfigCondition(ConditionProgressing, ConditionFalse, CheckImmediately, CheckDoNotRepeat)

		// Check that all requested components have been deployed
		CheckComponentsDeployment(components)
	} else {
		// Wait until Available condition is reported. Should be fast when no components are
		// being deployed
		CheckConfigCondition(ConditionAvailable, ConditionTrue, 5*time.Minute, CheckDoNotRepeat)
	}
}
