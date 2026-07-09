package types

type Label uint8

const (
	Label_Indirect Label = iota
	//
	Label_Dev
	Label_Optional
	Label_Root
	//
	Label_Placeholder
	Label_NonExistant
	//
	Label_ComponentType_Application
	Label_ComponentType_Container
	Label_ComponentType_CryptographicAsset
	Label_ComponentType_Data
	Label_ComponentType_Device
	Label_ComponentType_DeviceDriver
	Label_ComponentType_File
	Label_ComponentType_Firmware
	Label_ComponentType_Framework
	Label_ComponentType_Library
	Label_ComponentType_MachineLearningModel
	Label_ComponentType_OS
	Label_ComponentType_Platform
)
