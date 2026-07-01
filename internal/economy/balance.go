package economy

import "github.com/codyconfer/goose/internal/content"

var (
	costGrowth = content.Balance.CostGrowth

	BasePrice        = content.Balance.BasePrice
	consumerAppetite = content.Balance.ConsumerAppetite
	crowdHeadroom    = content.Balance.CrowdHeadroom
	crowdAdjustRate  = content.Balance.CrowdAdjustRate

	priceHoardOnOffer       = content.Balance.PriceHoardOnOffer
	priceMinSupply          = content.Balance.PriceMinSupply
	priceAdjustRate         = content.Balance.PriceAdjustRate
	priceVolatility         = content.Balance.PriceVolatility
	priceTrendDecay         = content.Balance.PriceTrendDecay
	priceTrendVol           = content.Balance.PriceTrendVol
	priceTrendMax           = content.Balance.PriceTrendMax
	priceShockTrend         = content.Balance.PriceShockTrend
	priceTrendTurbulence    = content.Balance.PriceTrendTurbulence
	priceTrendReversionDrag = content.Balance.PriceTrendReversionDrag
	priceCashflowTrend      = content.Balance.PriceCashflowTrend
	priceTradeTrend         = content.Balance.PriceTradeTrend
	priceDerivativeTrend    = content.Balance.PriceDerivativeTrend
	priceFloor              = content.Balance.PriceFloor
	priceCeil               = content.Balance.PriceCeil

	crierBonusPerLevel = content.Balance.CrierBonusPerLevel
	blitzBonusPerLevel = content.Balance.BlitzBonusPerLevel

	tradeFloorRate = content.Balance.TradeFloorRate
	tradeEpsilon   = content.Balance.TradeEpsilon

	SpecUnlockLevel       = content.Balance.SpecUnlockLevel
	SpecExpirySeconds     = content.Balance.SpecExpirySeconds
	SpecMarginPenalty     = content.Balance.SpecMarginPenalty
	SpecMaintenanceMargin = content.Balance.SpecMaintenanceMargin
	SpecLeverages         = content.Balance.SpecLeverages
	SpecPremiums          = content.Balance.SpecPremiums

	decommissionRefund = content.Balance.DecommissionRefund
	ledgerMax          = content.Balance.LedgerMax

	maxOfflineSeconds = content.Balance.MaxOfflineSeconds
)
