/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package config

const (
	BASE_URL_PARAM         = "BASE_URL"
	CLIENT_ID_PARAM        = "CLIENT_ID"
	CLIENT_SECRET_PARAM    = "CLIENT_SECRET"
	CERTIFICATE_PATH_PARAM = "CERT_PATH"
	PRODUCT_MODE_PARAM     = "PRODUCT_MODE"
)

// Deprecated constants for backward compatibility
const (
	ASGARDEO_BASE_URL_PARAM      = "ASGARDEO_BASE_URL"
	ASGARDEO_CLIENT_ID_PARAM     = "ASGARDEO_CLIENT_ID"
	ASGARDEO_CLIENT_SECRET_PARAM = "ASGARDEO_CLIENT_SECRET"
)

var ProductModes = struct {
	WSO2IS   string
	Asgardeo string
}{
	WSO2IS:   "wso2is",
	Asgardeo: "asgardeo",
}

var ProductNames = struct {
	WSO2IS   string
	Asgardeo string
}{
	WSO2IS:   "WSO2 Identity Server",
	Asgardeo: "Asgardeo",
}
