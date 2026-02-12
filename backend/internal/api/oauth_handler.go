package api

import (
	"api-aggregator/backend/internal/accountpool"
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	credRepo *repository.AccountCredentialRepository
	poolRepo *repository.AccountPoolRepository
}

func NewOAuthHandler(
	credRepo *repository.AccountCredentialRepository,
	poolRepo *repository.AccountPoolRepository,
) *OAuthHandler {
	return &OAuthHandler{
		credRepo: credRepo,
		poolRepo: poolRepo,
	}
}

// InitiateOAuth 启动 OAuth 流程
func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
	var req struct {
		PoolID            uint   `json:"pool_id" binding:"required"`
		Provider          string `json:"provider" binding:"required"`
		AuthType          string `json:"auth_type"` // "oauth", "device_code", "builder_id"
		Region            string `json:"region"`
		BuilderIDStartURL string `json:"builder_id_start_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证池存在
	pool, err := h.poolRepo.FindByID(c.Request.Context(), req.PoolID)
	if err != nil || pool == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pool not found"})
		return
	}

	// 验证提供商匹配
	if req.Provider != pool.Provider {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider mismatch"})
		return
	}

	// 获取提供商
	provider, err := accountpool.Get(req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 根据认证类型处理
	if req.AuthType == "device_code" || req.AuthType == "builder_id" {
		// 设备码流程
		deviceProvider, ok := provider.(accountpool.DeviceCodeProvider)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Provider does not support device code flow"})
			return
		}

		result, err := deviceProvider.InitiateDeviceCode(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 启动后台轮询任务
		go h.pollDeviceCodeInBackground(c.Request.Context(), req.PoolID, req.Provider, result)

		// 返回设备码信息给前端
		c.JSON(http.StatusOK, gin.H{
			"authUrl": result["verification_uri_complete"],
			"authInfo": gin.H{
				"device_code":                result["device_code"],
				"user_code":                  result["user_code"],
				"verification_uri":           result["verification_uri"],
				"verification_uri_complete":  result["verification_uri_complete"],
				"expires_in":                 result["expires_in"],
				"interval":                   result["interval"],
			},
		})
	} else {
		// OAuth 流程
		oauthProvider, ok := provider.(accountpool.OAuthProvider)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Provider does not support OAuth"})
			return
		}

		authURL, err := oauthProvider.GetAuthURL(c.Request.Context(), fmt.Sprintf("%d", req.PoolID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
	}
}

// OAuthCallback OAuth 回调处理
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	providerName := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state"})
		return
	}

	// 解析池 ID
	poolID, err := strconv.ParseUint(state, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// 验证池存在
	pool, err := h.poolRepo.FindByID(c.Request.Context(), uint(poolID))
	if err != nil || pool == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pool not found"})
		return
	}

	// 获取提供商
	provider, err := accountpool.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oauthProvider, ok := provider.(accountpool.OAuthProvider)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider does not support OAuth"})
		return
	}

	// 交换授权码
	cred, err := oauthProvider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置池 ID 和提供商
	cred.PoolID = uint(poolID)
	cred.Provider = providerName
	cred.IsActive = true

	// 保存凭据
	if _, err := h.credRepo.Create(c.Request.Context(), cred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "oauth_success.html", gin.H{
		"provider": providerName,
		"pool_id":  poolID,
	})
}

// PollDeviceCode 轮询设备码状态
func (h *OAuthHandler) PollDeviceCode(c *gin.Context) {
	var req struct {
		PoolID     uint   `json:"pool_id" binding:"required"`
		Provider   string `json:"provider" binding:"required"`
		DeviceCode string `json:"device_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证池存在
	pool, err := h.poolRepo.FindByID(c.Request.Context(), req.PoolID)
	if err != nil || pool == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pool not found"})
		return
	}

	// 获取提供商
	provider, err := accountpool.Get(req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceProvider, ok := provider.(accountpool.DeviceCodeProvider)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider does not support device code flow"})
		return
	}

	// 轮询设备码
	cred, err := deviceProvider.PollDeviceCode(c.Request.Context(), req.DeviceCode)
	if err != nil {
		if err.Error() == "authorization_pending" {
			c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置池 ID 和提供商
	cred.PoolID = req.PoolID
	cred.Provider = req.Provider
	cred.IsActive = true

	// 保存凭据
	savedCred, err := h.credRepo.Create(c.Request.Context(), cred)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, savedCred)
}

// pollDeviceCodeInBackground 后台轮询设备码
func (h *OAuthHandler) pollDeviceCodeInBackground(ctx context.Context, poolID uint, providerName string, deviceInfo map[string]interface{}) {
	deviceCode := deviceInfo["device_code"].(string)
	clientID := deviceInfo["client_id"].(string)
	clientSecret := deviceInfo["client_secret"].(string)
	region := deviceInfo["region"].(string)
	interval := deviceInfo["interval"].(int)
	expiresIn := deviceInfo["expires_in"].(int)
	
	maxAttempts := expiresIn / interval
	ssoOIDCEndpoint := fmt.Sprintf("https://oidc.%s.amazonaws.com", region)
	
	client := &http.Client{Timeout: 30 * time.Second}
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(time.Duration(interval) * time.Second)
		
		// 轮询 token
		reqBody, _ := json.Marshal(map[string]string{
			"clientId":     clientID,
			"clientSecret": clientSecret,
			"deviceCode":   deviceCode,
			"grantType":    "urn:ietf:params:oauth:grant-type:device_code",
		})
		
		req, err := http.NewRequestWithContext(ctx, "POST", ssoOIDCEndpoint+"/token", bytes.NewBuffer(reqBody))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "KiroIDE")
		
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		
		var result struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			ExpiresIn    int    `json:"expiresIn"`
			Error        string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		
		if result.Error == "authorization_pending" {
			continue
		}
		
		if result.Error == "slow_down" {
			time.Sleep(5 * time.Second)
			continue
		}
		
		if result.AccessToken != "" {
			// 成功获取 token，保存凭据
			expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
			cred := &models.AccountCredential{
				Name:     fmt.Sprintf("Kiro Builder ID - %s", time.Now().Format("2006-01-02 15:04")),
				Provider: providerName,
				AuthType: "builder_id",
				CredentialsData: models.CredentialsData{
					"access_token":  result.AccessToken,
					"refresh_token": result.RefreshToken,
					"client_id":     clientID,
					"client_secret": clientSecret,
					"region":        region,
					"auth_method":   "builder-id",
					"idc_region":    region,
				},
				ExpiresAt: &expiresAt,
				IsActive:  true,
				PoolID:    poolID,
			}
			
			// 创建凭据
			createdCred, err := h.credRepo.Create(ctx, cred)
			if err != nil {
				return
			}
			
			// 添加到账号池关联表
			h.poolRepo.AddCredentialToPool(ctx, poolID, createdCred.ID)
			return
		}
		
		// 其他错误，停止轮询
		return
	}
}
