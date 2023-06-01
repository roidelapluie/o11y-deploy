// Copyright 2023 The O11y Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alecthomas/kingpin/v2"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Define flags
	port := kingpin.Flag("port", "Port to listen on.").Default("9000").String()
	dirPath := kingpin.Flag("dir", "Path to directory to serve.").Default("ui/build").String()
	devMode := kingpin.Flag("dev", "Use development mode.").Default("false").Bool()

	// Parse the flags
	kingpin.Parse()

	// Create a new gin router
	r := gin.Default()

	// Create a new Prometheus registry
	promRegistry := gin.New()
	promRegistry.Use(gin.Logger(), gin.Recovery())
	promRegistry.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Any("/metrics", promRegistry.HandleContext)

	// If in development mode, proxy requests to localhost:3000
	if *devMode {
		proxyURL, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)

		r.Any("/ui/*path", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	} else {
		// Serve the directory at the /ui path
		r.StaticFS("/ui", http.Dir(*dirPath))
		// Redirect to /ui when accessing the root path
	}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui")
	})

	// Start the server
	r.Run(":" + *port)
}
