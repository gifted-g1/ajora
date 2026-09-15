package domain

type AjoStatus string

const (
    AjoDraft     AjoStatus = "DRAFT"
    AjoOpen      AjoStatus = "OPEN"
    AjoActive    AjoStatus = "ACTIVE"
    AjoCompleted AjoStatus = "COMPLETED"
    AjoCancelled AjoStatus = "CANCELLED"
)

type RoundStatus string

const (
    RoundUpcoming       RoundStatus = "UPCOMING"
    RoundContributing   RoundStatus = "CONTRIBUTING"
    RoundComplete       RoundStatus = "ROUND_COMPLETE"
    RoundPayoutReady    RoundStatus = "PAYOUT_READY"
    RoundPayoutComplete RoundStatus = "PAYOUT_COMPLETED"
)

type ContributionStatus string

const (
    ContributionPending    ContributionStatus = "PENDING"
    ContributionProcessing ContributionStatus = "PROCESSING"
    ContributionConfirmed  ContributionStatus = "CONFIRMED"
    ContributionLate       ContributionStatus = "LATE"
    ContributionFailed     ContributionStatus = "FAILED"
)

type PayoutStatus string

const (
    PayoutScheduled  PayoutStatus = "SCHEDULED"
    PayoutReady      PayoutStatus = "READY"
    PayoutProcessing PayoutStatus = "PROCESSING"
    PayoutCompleted  PayoutStatus = "COMPLETED"
    PayoutFailed     PayoutStatus = "FAILED"
)

type TransactionType string

const (
    TxDeposit         TransactionType = "DEPOSIT"
    TxWithdrawal      TransactionType = "WITHDRAWAL"
    TxAjoContribution TransactionType = "AJO_CONTRIBUTION"
    TxAjoPayout       TransactionType = "AJO_PAYOUT"
    TxLateFee         TransactionType = "LATE_FEE"
)

type Frequency string

const (
    FreqDaily    Frequency = "DAILY"
    FreqWeekly   Frequency = "WEEKLY"
    FreqBiweekly Frequency = "BIWEEKLY"
    FreqMonthly  Frequency = "MONTHLY"
)

type PayoutOrderStrategy string

const (
    PayoutFixed         PayoutOrderStrategy = "FIXED"
    PayoutFirstCome     PayoutOrderStrategy = "FIRST_COME"
    PayoutRandomLottery PayoutOrderStrategy = "RANDOM_LOTTERY"
)
