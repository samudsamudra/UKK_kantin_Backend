package app

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//
// =========================
// ENUMS
// =========================
//

type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdminStan  UserRole = "admin_stan"
	RoleSiswa      UserRole = "siswa"
)

type MenuJenis string

const (
	JenisMakanan MenuJenis = "makanan"
	JenisMinuman MenuJenis = "minuman"
)

type TransaksiStatus string

const (
	StatusBelumDikonfirm TransaksiStatus = "belum dikonfirm"
	StatusDimasak        TransaksiStatus = "dimasak"
	StatusDiantar        TransaksiStatus = "diantar"
	StatusSampai         TransaksiStatus = "sampai"
)

//
// =========================
// USER (BASE IDENTITY)
// =========================
//

type User struct {
	ID                 uint      `gorm:"primaryKey" json:"-"`
	PublicID           string    `gorm:"size:36;uniqueIndex;not null" json:"user_id"`
	Email              string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	PasswordHash       string    `gorm:"size:255;not null" json:"-"`
	Role               UserRole  `gorm:"size:50;not null" json:"role"`
	MustChangePassword bool      `gorm:"default:true" json:"must_change_password"`
	CreatedBy          *uint     `gorm:"index" json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Saldo     float64   `gorm:"type:decimal(15,2);default:0"`

	Siswa *Siswa `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID"`

	Stan  *Stan  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.PublicID == "" {
		u.PublicID = uuid.NewString()
	}
	return nil
}

//
// =========================
// SISWA
// =========================
//

type Siswa struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	PublicID  string    `gorm:"size:36;uniqueIndex;not null" json:"siswa_id"`
	Nama      string    `gorm:"size:150;not null" json:"nama_lengkap"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Siswa) BeforeCreate(tx *gorm.DB) error {
	if s.PublicID == "" {
		s.PublicID = uuid.NewString()
	}
	return nil
	
}

//
// =========================
// STAN (ADMIN STAN)
// =========================
//

type Stan struct {
	ID          uint      `gorm:"primaryKey"`
	PublicID    string    `gorm:"size:36;uniqueIndex"`
	NamaStan    string    `gorm:"size:100"`
	NamaPemilik string    `gorm:"size:100"`
	Telp        string    `gorm:"size:20"`
	UserID      uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}


func (s *Stan) BeforeCreate(tx *gorm.DB) error {
	if s.PublicID == "" {
		s.PublicID = uuid.NewString()
	}
	return nil
}

//
// =========================
// MENU
// =========================
//

type Menu struct {
	ID          uint      `gorm:"primaryKey"`
	PublicID    string    `gorm:"size:36;uniqueIndex"`
	NamaMakanan string    `gorm:"size:100"`
	Harga       float64
	Jenis       MenuJenis
	Deskripsi   string
	StanID      uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}


func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.PublicID == "" {
		m.PublicID = uuid.NewString()
	}
	return nil
}

//
// =========================
// DISKON
// =========================
//

type Diskon struct {
	ID          uint       `gorm:"primaryKey" json:"-"`
	PublicID    string     `gorm:"size:36;uniqueIndex;not null" json:"diskon_id"`

	StanID      uint       `gorm:"index;not null"`
	Nama        string     `gorm:"size:100;not null"`
	Persentase  float64    `gorm:"not null"`

	TanggalAwal  *time.Time `gorm:"index"`
	TanggalAkhir *time.Time `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d *Diskon) BeforeCreate(tx *gorm.DB) error {
	if d.PublicID == "" {
		d.PublicID = uuid.NewString()
	}
	return nil
}

//
// =========================
// TRANSAKSI
// =========================
//

type Transaksi struct {
	ID        uint            `gorm:"primaryKey" json:"-"`
	PublicID  string          `gorm:"size:36;uniqueIndex;not null" json:"transaksi_id"`
	StanID    uint            `gorm:"index;not null" json:"-"`
	SiswaID   uint            `gorm:"index;not null" json:"-"`
	Status    TransaksiStatus `gorm:"size:50;not null" json:"status"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Details []DetailTransaksi `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:TransaksiID"`
}

func (t *Transaksi) BeforeCreate(tx *gorm.DB) error {
	if t.PublicID == "" {
		t.PublicID = uuid.NewString()
	}
	return nil
}

//
// =========================
// DETAIL TRANSAKSI
// =========================
//

type DetailTransaksi struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	TransaksiID uint      `gorm:"index;not null" json:"-"`
	MenuID      uint      `gorm:"index;not null" json:"-"`
	Qty         int       `gorm:"not null" json:"qty"`
	HargaBeli   float64   `gorm:"not null" json:"harga_beli"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Menu Menu `gorm:"foreignKey:MenuID"`
}

//
// =========================
// WALLET (OPTIONAL / FUTURE)
// =========================
//

type WalletTransaction struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	PublicID  string    `gorm:"size:36;uniqueIndex;not null" json:"wallet_tx_id"`
	UserID    uint      `gorm:"index;not null" json:"-"`
	Amount    float64   `gorm:"not null" json:"amount"`
	Type      string    `gorm:"size:50;not null" json:"type"` // topup | debit
	Note      string    `gorm:"type:text" json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (w *WalletTransaction) BeforeCreate(tx *gorm.DB) error {
	if w.PublicID == "" {
		w.PublicID = uuid.NewString()
	}
	return nil
}
