package topics

const (
	AccountsDebitAdviceSuccess = "com.eyowo.accounts.debit.advice.success"
	AccountsDebitAdviceFailed  = "com.eyowo.accounts.debit.advice.failed"

	AccountsCreditAdviceSuccess = "com.eyowo.accounts.credit.advice.success"
	AccountsCreditAdviceFailed  = "com.eyowo.accounts.credit.advice.failed"
)

const (
	StaticAccountCreated           = "com.eyowo.accounts.static.account.created"
	StaticAccountUpdated           = "com.eyowo.accounts.static.account.updated"
	StaticAccountTierUpgraded      = "com.eyowo.accounts.static.account.tier.updated"
	StaticAccountStatusActivated   = "com.eyowo.accounts.static.account.activated"
	StaticAccountStatusDeactivated = "com.eyowo.accounts.static.account.deactivated"

	StaticAccountPinSet     = "com.eyowo.accounts.static.account.pin.set"
	StaticAccountPinChanged = "com.eyowo.accounts.static.account.pin.changed"

	StaticAccountDailyTransactionLimitUpdated = "com.eyowo.accounts.static.account.transaction.daily.limit.updated"
	StaticAccountUnitTransactionLimitUpdated  = "com.eyowo.accounts.static.account.transaction.unit.limit.updated"

	StaticAccountLienLocked   = "com.eyowo.accounts.static.account.lien.locked"
	StaticAccountLienUnlocked = "com.eyowo.accounts.static.account.lien.unlocked"
)

const (
	DynamicAccountCreated   = "com.eyowo.accounts.dynamic.account.created"
	DynamicAccountDestroyed = "com.eyowo.accounts.dynamic.account.destroyed"
)

const (
	TransferDebitAdvice  = "com.eyowo.transfers.debit.advice"
	TransferCreditAdvice = "com.eyowo.transfers.credit.advice"
)

const (
	KYCUserSelfVerificationUpdated     = "com.eyowo.kyc.user.self.verification.updated"
	KYCUserBVNVerificationUpdated      = "com.eyowo.kyc.user.bvn.verification.updated"
	KYCUserIdentityVerificationUpdated = "com.eyowo.kyc.user.identity.verification.updated"
	KYCBusinessVerificationUpdated     = "com.eyowo.kyc.business.verification.updated"
)
