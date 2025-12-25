package scryfall

// Card represents the subset of the Scryfall card schema we ingest.
type Card struct {
	ID              string            `json:"id"`
	OracleID        string            `json:"oracle_id"`
	Name            string            `json:"name"`
	Lang            string            `json:"lang"`
	ReleasedAt      string            `json:"released_at"`
	Set             string            `json:"set"`
	CollectorNumber string            `json:"collector_number"`
	Rarity          string            `json:"rarity"`
	Layout          string            `json:"layout"`
	ManaCost        string            `json:"mana_cost"`
	TypeLine        string            `json:"type_line"`
	OracleText      string            `json:"oracle_text"`
	Power           string            `json:"power"`
	Toughness       string            `json:"toughness"`
	Loyalty         string            `json:"loyalty"`
	CMC             float64           `json:"cmc"`
	Keywords        []string          `json:"keywords"`
	Prices          CardPrices        `json:"prices"`
	ImageURIs       map[string]string `json:"image_uris"`
	CardFaces       []CardFace        `json:"card_faces"`
	TcgplayerID     int               `json:"tcgplayer_id"`
	CardmarketID    int               `json:"cardmarket_id"`
	Uri             string            `json:"uri"`
	ScryfallURI     string            `json:"scryfall_uri"`
	RulingsURI      string            `json:"rulings_uri"`
	PrintsSearchURI string            `json:"prints_search_uri"`
	Digital         bool              `json:"digital"`
	Reserved        bool              `json:"reserved"`
	EDHRecRank      int               `json:"edhrec_rank"`
	PennyRank       int               `json:"penny_rank"`
	Games           []string          `json:"games"`
	Promo           bool              `json:"promo"`
	Reprint         bool              `json:"reprint"`
	Variation       bool              `json:"variation"`
	Oversized       bool              `json:"oversized"`
	StorySpotlight  bool              `json:"story_spotlight"`
	FullArt         bool              `json:"full_art"`
	Textless        bool              `json:"textless"`
	Booster         bool              `json:"booster"`
	FrameEffects    []string          `json:"frame_effects"`
	Frame           string            `json:"frame"`
	SecurityStamp   string            `json:"security_stamp"`
	BorderColor     string            `json:"border_color"`
	Watermark       string            `json:"watermark"`
}

// CardPrices represent the various market prices returned by Scryfall as strings.
type CardPrices struct {
	USD       string `json:"usd"`
	USDFoil   string `json:"usd_foil"`
	USDEtched string `json:"usd_etched"`
	EUR       string `json:"eur"`
	EURFoil   string `json:"eur_foil"`
	TIX       string `json:"tix"`
}

// CardFace captures the data returned for double-faced cards.
type CardFace struct {
	Name       string            `json:"name"`
	ManaCost   string            `json:"mana_cost"`
	TypeLine   string            `json:"type_line"`
	OracleText string            `json:"oracle_text"`
	Colors     []string          `json:"colors"`
	Power      string            `json:"power"`
	Toughness  string            `json:"toughness"`
	Loyalty    string            `json:"loyalty"`
	FlavorText string            `json:"flavor_text"`
	ImageURIs  map[string]string `json:"image_uris"`
}

// CardBulkData describes downloadable data sets available from Scryfall.
type CardBulkData struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	UpdatedAt       string `json:"updated_at"`
	URI             string `json:"uri"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DownloadURI     string `json:"download_uri"`
	ContentType     string `json:"content_type"`
	ContentEncoding string `json:"content_encoding"`
	CompressedSize  int64  `json:"compressed_size"`
	PermalinkURI    string `json:"permalink_uri"`
}

// CardSet represents a Scryfall set object.
type CardSet struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	ReleasedAt  string `json:"released_at"`
	SetType     string `json:"set_type"`
	CardCount   int    `json:"card_count"`
	Digital     bool   `json:"digital"`
	NonfoilOnly bool   `json:"nonfoil_only"`
	FoilOnly    bool   `json:"foil_only"`
	IconSVGURI  string `json:"icon_svg_uri"`
}
