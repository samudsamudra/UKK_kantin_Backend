package app

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- enums ---

type UserRole string

const (
	RoleSiswa     UserRole = "siswa"
	RoleAdminStan UserRole = "admin_stan"
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

// --- USER ---

type User struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	PublicID  string    `gorm:"size:36;uniqueIndex;not null" json:"user_id"`
	Username  string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      UserRole  `gorm:"size:50;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Siswa *Siswa `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UserID"`
	Stan  *Stan  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.PublicID == "" {
		u.PublicID = uuid.NewString()
	}
	return nil
}

// --- SISWA ---

type Siswa struct {
	ID        uint   `gorm:"primaryKey" json:"-"`
	PublicID  string `gorm:"size:36;uniqueIndex;not null" json:"siswa_id"`
	Nama      string `gorm:"size:100;not null" json:"nama_siswa"`
	Alamat    string `gorm:"type:text" json:"alamat"`
	Telp      string `gorm:"size:20" json:"telp"`
	Foto      string `gorm:"size:255" json:"foto"`
	UserID    *uint  `gorm:"uniqueIndex" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Siswa) BeforeCreate(tx *gorm.DB) (err error) {
	if s.PublicID == "" {
		s.PublicID = uuid.NewString()
	}
	return nil
}

// --- STAN ADMIN ---

type Stan struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	PublicID    string `gorm:"size:36;uniqueIndex;not null" json:"stan_id"`
	NamaStan    string `gorm:"size:100;not null" json:"nama_stan"`
	NamaPemilik string `gorm:"size:100" json:"nama_pemilik"`
	Telp        string `gorm:"size:20" json:"telp"`
	UserID      *uint  `gorm:"uniqueIndex" json:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Menus []Menu `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:StanID"`
}

func (s *Stan) BeforeCreate(tx *gorm.DB) (err error) {
	if s.PublicID == "" {
		s.PublicID = uuid.NewString()
	}
	return nil
}

// --- MENU ---

type Menu struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	PublicID    string    `gorm:"size:36;uniqueIndex;not null" json:"menu_id"`
	NamaMakanan string    `gorm:"size:100;not null" json:"nama_makanan"`
	Harga       float64   `json:"harga"`
	Jenis       MenuJenis `gorm:"size:50;not null" json:"jenis"`
	Foto        string    `gorm:"size:255" json:"foto"`
	Deskripsi   string    `gorm:"type:text" json:"deskripsi"`
	StanID      *uint     `gorm:"index" json:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Diskons []Diskon `gorm:"many2many:menu_diskons;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (m *Menu) BeforeCreate(tx *gorm.DB) (err error) {
	if m.PublicID == "" {
		m.PublicID = uuid.NewString()
	}
	return nil
}

// --- DISKON ---

// internal/app/models.go (potongan)
type Diskon struct {
	ID               uint       `gorm:"primaryKey" json:"-"`
	PublicID         string     `gorm:"size:36;uniqueIndex;not null" json:"diskon_id"`
	NamaDiskon       string     `gorm:"size:100;not null" json:"nama_diskon"`
	PersentaseDiskon float64    `json:"persentase_diskon"`                                     // 0..100
	TanggalAwal      *time.Time `gorm:"index:idx_diskon_time,priority:1" json:"tanggal_awal"`  // nil = always active (opsional)
	TanggalAkhir     *time.Time `gorm:"index:idx_diskon_time,priority:2" json:"tanggal_akhir"` // nil = no end
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Menus []Menu `gorm:"many2many:menu_diskons;"`
}

func (d *Diskon) BeforeCreate(tx *gorm.DB) (err error) {
	if d.PublicID == "" {
		d.PublicID = uuid.NewString()
	}
	return nil
}

// --- JOIN TABLE ---

type MenuDiskon struct {
	MenuID   uint `gorm:"primaryKey"`
	DiskonID uint `gorm:"primaryKey"`
}

// --- TRANSAKSI ---

type Transaksi struct {
	ID        uint            `gorm:"primaryKey" json:"-"`
	PublicID  string          `gorm:"size:36;uniqueIndex;not null" json:"transaksi_id"`
	Tanggal   time.Time       `gorm:"not null;index" json:"tanggal"`
	StanID    uint            `gorm:"index;not null" json:"-"`
	SiswaID   uint            `gorm:"index;not null" json:"-"`
	Status    TransaksiStatus `gorm:"size:50;not null" json:"status"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Details []DetailTransaksi `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:TransaksiID"`
}

func (t *Transaksi) BeforeCreate(tx *gorm.DB) (err error) {
	if t.PublicID == "" {
		t.PublicID = uuid.NewString()
	}
	if t.Tanggal.IsZero() {
		t.Tanggal = time.Now()
	}
	return nil
}

// --- DETAIL TRANSAKSI ---

type DetailTransaksi struct {
	ID          uint    `gorm:"primaryKey" json:"-"`
	TransaksiID uint    `gorm:"index;not null" json:"-"`
	MenuID      uint    `gorm:"index;not null" json:"-"`
	Qty         int     `gorm:"not null" json:"qty"`
	HargaBeli   float64 `gorm:"not null" json:"harga_beli"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Menu Menu `gorm:"foreignKey:MenuID"`
}
