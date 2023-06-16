# HwInventory100HwInventory

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**XName** | **string** |  | [optional] [default to null]
**Format** | **string** | How results are displayed   FullyFlat      All component types listed in their own                  arrays only.  No nesting of any children   Hierarchical   All subcomponents listed as children up to                  top level component (or set of cabinets)   NestNodesOnly  Flat except that node subcomponents are nested                  hierarchically. Default is NestNodesOnly. | [optional] [default to null]
**Cabinets** | [**[]HwInvByLocCabinet**](HWInvByLocCabinet.md) | All components with HMS type &#x27;Cabinet&#x27; appropriate given Target component/partition and query type. | [optional] [default to null]
**Chassis** | [**[]HwInvByLocChassis**](HWInvByLocChassis.md) | All appropriate components with HMS type &#x27;Chassis&#x27; given Target component/partition and query type. | [optional] [default to null]
**ComputeModules** | [**[]HwInvByLocComputeModule**](HWInvByLocComputeModule.md) | All appropriate components with HMS type &#x27;ComputeModule&#x27; given Target component/partition and query type. | [optional] [default to null]
**RouterModules** | [**[]HwInvByLocRouterModule**](HWInvByLocRouterModule.md) | All appropriate components with HMS type &#x27;RouterModule&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeEnclosures** | [**[]HwInvByLocNodeEnclosure**](HWInvByLocNodeEnclosure.md) | All appropriate components with HMS type &#x27;NodeEnclosure&#x27; given Target component/partition and query type. | [optional] [default to null]
**HSNBoards** | [**[]HwInvByLocHsnBoard**](HWInvByLocHSNBoard.md) | All appropriate components with HMS type &#x27;HSNBoard&#x27; given Target component/partition and query type. | [optional] [default to null]
**MgmtSwitches** | [**[]HwInvByLocMgmtSwitch**](HWInvByLocMgmtSwitch.md) | All appropriate components with HMS type &#x27;MgmtSwitch&#x27; given Target component/partition and query type. | [optional] [default to null]
**MgmtHLSwitches** | [**[]HwInvByLocMgmtHlSwitch**](HWInvByLocMgmtHLSwitch.md) | All appropriate components with HMS type &#x27;MgmtHLSwitch&#x27; given Target component/partition and query type. | [optional] [default to null]
**CDUMgmtSwitches** | [**[]HwInvByLocCduMgmtSwitch**](HWInvByLocCDUMgmtSwitch.md) | All appropriate components with HMS type &#x27;CDUMgmtSwitch&#x27; given Target component/partition and query type. | [optional] [default to null]
**Nodes** | [**[]HwInvByLocNode**](HWInvByLocNode.md) | All appropriate components with HMS type &#x27;Node&#x27; given Target component/partition and query type. | [optional] [default to null]
**Processors** | [**[]HwInvByLocProcessor**](HWInvByLocProcessor.md) | All appropriate components with HMS type &#x27;Processor&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeAccels** | [**[]HwInvByLocNodeAccel**](HWInvByLocNodeAccel.md) | All appropriate components with HMS type &#x27;NodeAccel&#x27; given Target component/partition and query type. | [optional] [default to null]
**Drives** | [**[]HwInvByLocDrive**](HWInvByLocDrive.md) | All appropriate components with HMS type &#x27;Drive&#x27; given Target component/partition and query type. | [optional] [default to null]
**Memory** | [**[]HwInvByLocMemory**](HWInvByLocMemory.md) | All appropriate components with HMS type &#x27;Memory&#x27; given Target component/partition and query type. | [optional] [default to null]
**CabinetPDUs** | [**[]HwInvByLocPdu**](HWInvByLocPDU.md) | All appropriate components with HMS type &#x27;CabinetPDU&#x27; given Target component/partition and query type. | [optional] [default to null]
**CabinetPDUPowerConnectors** | [**[]HwInvByLocOutlet**](HWInvByLocOutlet.md) | All appropriate components with HMS type &#x27;CabinetPDUPowerConnector&#x27; given Target component/partition and query type. | [optional] [default to null]
**CMMRectifiers** | [**[]HwInvByLocCmmRectifier**](HWInvByLocCMMRectifier.md) | All appropriate components with HMS type &#x27;CMMRectifier&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeAccelRisers** | [**[]HwInvByLocNodeAccelRiser**](HWInvByLocNodeAccelRiser.md) | All appropriate components with HMS type &#x27;NodeAccelRiser&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeHsnNICs** | [**[]HwInvByLocHsnnic**](HWInvByLocHSNNIC.md) | All appropriate components with HMS type &#x27;NodeHsnNic&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeEnclosurePowerSupplies** | [**[]HwInvByLocNodeEnclosurePowerSupply**](HWInvByLocNodeEnclosurePowerSupply.md) | All appropriate components with HMS type &#x27;NodeEnclosurePowerSupply&#x27; given Target component/partition and query type. | [optional] [default to null]
**NodeBMC** | [**[]HwInvByLocNodeBmc**](HWInvByLocNodeBMC.md) | All appropriate components with HMS type &#x27;NodeBMC&#x27; given Target component/partition and query type. | [optional] [default to null]
**RouterBMC** | [**[]HwInvByLocRouterBmc**](HWInvByLocRouterBMC.md) | All appropriate components with HMS type &#x27;RouterBMC&#x27; given Target component/partition and query type. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

