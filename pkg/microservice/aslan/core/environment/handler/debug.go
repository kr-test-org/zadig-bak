/*
Copyright 2022 The KodeRover Authors.

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

package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/koderover/zadig/pkg/microservice/aslan/core/environment/service"
	"github.com/koderover/zadig/pkg/setting"
	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
	"github.com/koderover/zadig/pkg/types"
)

func PatchDebugContainer(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Param("env")
	podName := c.Param("podName")
	projectName := c.Query("projectName")
	debugImage := c.Query("debugImage")
	if debugImage == "" {
		debugImage = types.DebugImage
	}

	internalhandler.InsertDetailedOperationLog(c, ctx.UserName, projectName, setting.OperationSceneEnv, "启动调试容器", "环境", envName, "", ctx.Logger, envName)

	ctx.Err = service.PatchDebugContainer(c, projectName, envName, podName, debugImage)
}
