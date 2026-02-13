package pricing

import (
	"api-aggregator/backend/internal/domain/apiconfig"
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service 定价服务接口
type Service interface {
	CreatePricing(ctx context.Context, req *CreatePricingRequest) (*PricingResponse, error)
	GetPricing(ctx context.Context, id uint) (*PricingResponse, error)
	GetPricings(ctx context.Context, req *GetPricingsRequest) (*PricingListResponse, error)
	GetAllPricings(ctx context.Context) ([]*PricingResponse, error)
	GetActivePricings(ctx context.Context) ([]*PricingResponse, error)
	GetPricingsByAPIConfig(ctx context.Context, apiConfigID uint) ([]*PricingResponse, error)
	GetPricingsByModel(ctx context.Context, modelName string) ([]*PricingResponse, error)
	UpdatePricing(ctx context.Context, id uint, req *UpdatePricingRequest) (*PricingResponse, error)
	DeletePricing(ctx context.Context, id uint) error
	CalculateCost(ctx context.Context, req *CalculateCostRequest) (*CostCalculationResponse, error)
	BatchCreatePricings(ctx context.Context, req *BatchCreatePricingRequest) (*BatchCreatePricingResponse, error)
}

// service 定价服务实现
type service struct {
	repo            Repository
	apiConfigRepo   apiconfig.Repository
	logger          logger.Logger
}

// NewService 创建定价服务
func NewService(repo Repository, apiConfigRepo apiconfig.Repository, logger logger.Logger) Service {
	return &service{
		repo:          repo,
		apiConfigRepo: apiConfigRepo,
		logger:        logger,
	}
}

// CreatePricing 创建定价
func (s *service) CreatePricing(ctx context.Context, req *CreatePricingRequest) (*PricingResponse, error) {
	// 检查是否已存在
	existing, err := s.repo.FindByModelAndAPIConfig(ctx, req.ModelName, req.APIConfigID)
	if err != nil {
		s.logger.Error("Failed to check existing pricing",
			logger.String("model", req.ModelName),
			logger.Uint("api_config_id", req.APIConfigID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to check existing pricing")
	}
	
	if existing != nil {
		// 如果已存在，直接返回
		s.logger.Info("Pricing already exists, returning existing record",
			logger.Uint("pricing_id", existing.ID),
			logger.String("model", req.ModelName),
			logger.Uint("api_config_id", req.APIConfigID))
		return existing.ToResponse(), nil
	}

	// 设置默认值
	currency := req.Currency
	if currency == "" {
		currency = "credits"
	}
	unit := req.Unit
	if unit == 0 {
		unit = 1000
	}

	// 创建定价
	pricing := &Pricing{
		APIConfigID: req.APIConfigID,
		ModelName:   req.ModelName,
		InputPrice:  req.InputPrice,
		OutputPrice: req.OutputPrice,
		Currency:    currency,
		Unit:        unit,
		IsActive:    true,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, pricing); err != nil {
		s.logger.Error("Failed to create pricing",
			logger.String("model", req.ModelName),
			logger.Uint("api_config_id", req.APIConfigID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create pricing")
	}

	s.logger.Info("Pricing created successfully",
		logger.Uint("pricing_id", pricing.ID),
		logger.String("model", pricing.ModelName),
		logger.Uint("api_config_id", pricing.APIConfigID))

	return pricing.ToResponse(), nil
}

// GetPricing 获取定价
func (s *service) GetPricing(ctx context.Context, id uint) (*PricingResponse, error) {
	pricing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get pricing", logger.Uint("pricing_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get pricing")
	}
	if pricing == nil {
		return nil, errors.New(404001, "Pricing not found")
	}

	return pricing.ToResponse(), nil
}

// GetPricings 获取定价列表
func (s *service) GetPricings(ctx context.Context, req *GetPricingsRequest) (*PricingListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.APIConfigID > 0 {
		filters = append(filters, query.Filter{
			Field:    "api_config_id",
			Operator: "=",
			Value:    req.APIConfigID,
		})
	}
	if req.ModelName != "" {
		filters = append(filters, query.Filter{
			Field:    "model_name",
			Operator: "=",
			Value:    req.ModelName,
		})
	}
	if req.IsActive != nil {
		filters = append(filters, query.Filter{
			Field:    "is_active",
			Operator: "=",
			Value:    *req.IsActive,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "api_config_id", Desc: false},
		{Field: "model_name", Desc: false},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询定价列表
	pricings, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get pricings", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get pricings")
	}

	return &PricingListResponse{
		Pricings: s.toResponseListWithAPIConfig(ctx, pricings),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAllPricings 获取所有定价
func (s *service) GetAllPricings(ctx context.Context) ([]*PricingResponse, error) {
	pricings, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all pricings", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get all pricings")
	}

	return s.toResponseListWithAPIConfig(ctx, pricings), nil
}

// GetActivePricings 获取所有激活的定价
func (s *service) GetActivePricings(ctx context.Context) ([]*PricingResponse, error) {
	pricings, err := s.repo.FindActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get active pricings", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get active pricings")
	}

	return s.toResponseListWithAPIConfig(ctx, pricings), nil
}

// GetPricingsByAPIConfig 根据API配置获取定价
func (s *service) GetPricingsByAPIConfig(ctx context.Context, apiConfigID uint) ([]*PricingResponse, error) {
	pricings, err := s.repo.FindByAPIConfig(ctx, apiConfigID)
	if err != nil {
		s.logger.Error("Failed to get pricings by API config",
			logger.Uint("api_config_id", apiConfigID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get pricings by API config")
	}

	return s.toResponseListWithAPIConfig(ctx, pricings), nil
}

// GetPricingsByModel 根据模型获取定价
func (s *service) GetPricingsByModel(ctx context.Context, modelName string) ([]*PricingResponse, error) {
	pricings, err := s.repo.FindByModel(ctx, modelName)
	if err != nil {
		s.logger.Error("Failed to get pricings by model",
			logger.String("model", modelName),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get pricings by model")
	}

	return s.toResponseListWithAPIConfig(ctx, pricings), nil
}

// UpdatePricing 更新定价
func (s *service) UpdatePricing(ctx context.Context, id uint, req *UpdatePricingRequest) (*PricingResponse, error) {
	// 查找定价
	pricing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get pricing", logger.Uint("pricing_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get pricing")
	}
	if pricing == nil {
		return nil, errors.New(404001, "Pricing not found")
	}

	// 更新字段
	if req.InputPrice != nil {
		pricing.InputPrice = *req.InputPrice
	}
	if req.OutputPrice != nil {
		pricing.OutputPrice = *req.OutputPrice
	}
	if req.Currency != "" {
		pricing.Currency = req.Currency
	}
	if req.Unit != nil {
		pricing.Unit = *req.Unit
	}
	if req.IsActive != nil {
		pricing.IsActive = *req.IsActive
	}
	if req.Description != "" {
		pricing.Description = req.Description
	}

	// 保存更新
	if err := s.repo.Update(ctx, pricing); err != nil {
		s.logger.Error("Failed to update pricing",
			logger.Uint("pricing_id", id),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to update pricing")
	}

	s.logger.Info("Pricing updated successfully",
		logger.Uint("pricing_id", id),
		logger.String("model", pricing.ModelName))

	return pricing.ToResponse(), nil
}

// DeletePricing 删除定价
func (s *service) DeletePricing(ctx context.Context, id uint) error {
	// 检查定价是否存在
	pricing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get pricing", logger.Uint("pricing_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get pricing")
	}
	if pricing == nil {
		return errors.New(404001, "Pricing not found")
	}

	// 删除定价
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete pricing",
			logger.Uint("pricing_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete pricing")
	}

	s.logger.Info("Pricing deleted successfully",
		logger.Uint("pricing_id", id),
		logger.String("model", pricing.ModelName))

	return nil
}

// CalculateCost 计算成本
func (s *service) CalculateCost(ctx context.Context, req *CalculateCostRequest) (*CostCalculationResponse, error) {
	// 查找定价
	pricing, err := s.repo.FindByModelAndAPIConfig(ctx, req.ModelName, req.APIConfigID)
	if err != nil {
		s.logger.Error("Failed to find pricing",
			logger.String("model", req.ModelName),
			logger.Uint("api_config_id", req.APIConfigID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find pricing")
	}
	if pricing == nil {
		return nil, errors.New(404001, "Pricing not found for this model and API config")
	}

	// 计算成本
	inputCost := pricing.CalculateInputCost(req.InputTokens)
	outputCost := pricing.CalculateOutputCost(req.OutputTokens)
	totalCost := inputCost + outputCost

	return &CostCalculationResponse{
		ModelName:    req.ModelName,
		InputTokens:  req.InputTokens,
		OutputTokens: req.OutputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    totalCost,
		Currency:     pricing.Currency,
		Unit:         pricing.Unit,
	}, nil
}

// BatchCreatePricings 批量创建定价
func (s *service) BatchCreatePricings(ctx context.Context, req *BatchCreatePricingRequest) (*BatchCreatePricingResponse, error) {
	var created int
	var failed int
	var errorMessages []string

	for _, pricingReq := range req.Pricings {
		_, err := s.CreatePricing(ctx, &pricingReq)
		if err != nil {
			failed++
			errorMessages = append(errorMessages, err.Error())
		} else {
			created++
		}
	}

	s.logger.Info("Batch create pricings completed",
		logger.Int("created", created),
		logger.Int("failed", failed))

	return &BatchCreatePricingResponse{
		Created: created,
		Failed:  failed,
		Errors:  errorMessages,
	}, nil
}


// toResponseListWithAPIConfig 转换为响应列表并加载 API 配置信息
func (s *service) toResponseListWithAPIConfig(ctx context.Context, pricings []*Pricing) []*PricingResponse {
	if len(pricings) == 0 {
		return []*PricingResponse{}
	}

	// 收集所有唯一的 API 配置 ID
	apiConfigIDs := make(map[uint]bool)
	for _, p := range pricings {
		apiConfigIDs[p.APIConfigID] = true
	}

	// 批量加载 API 配置
	apiConfigMap := make(map[uint]*apiconfig.APIConfig)
	for id := range apiConfigIDs {
		config, err := s.apiConfigRepo.FindByID(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to load API config",
				logger.Uint("api_config_id", id),
				logger.Error(err))
			continue
		}
		if config != nil {
			apiConfigMap[id] = config
		}
	}

	// 转换为响应对象
	responses := make([]*PricingResponse, len(pricings))
	for i, pricing := range pricings {
		resp := pricing.ToResponse()
		
		// 添加 API 配置信息
		if config, ok := apiConfigMap[pricing.APIConfigID]; ok {
			resp.APIConfig = &APIConfigInfo{
				ID:   config.ID,
				Name: config.Name,
				Type: config.Type,
			}
		}
		
		responses[i] = resp
	}

	return responses
}
