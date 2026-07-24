package web

import "embed"

//go:embed static/*
var StaticEmbed embed.FS

////go:embed openapi/*
//var OpenAPIEmbed embed.FS
