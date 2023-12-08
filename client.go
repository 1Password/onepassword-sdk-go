package main

type ClientConfig struct {
	SAToken               string `json:"saToken"`
	Language              string `json:"language"`
	AppName               string `json:"appName"`
	AppVersion            string `json:"appVersion"`
	RequestLibraryName    string `json:"requestLibraryName"`
	RequestLibraryVersion string `json:"requestLibraryVersion"`
}
