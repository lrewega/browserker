//revive:disable
package crawler

import (
	"regexp"
)

// taken from https://source.chromium.org/chromium/chromium/src/+/master:components/autofill/core/common/autofill_regex_constants.cc?originalUrl=https:%2F%2Fcs.chromium.org%2F
var DateStartRe = regexp.MustCompile("start|depart|leave|begin|from")
var DateEndRe = regexp.MustCompile("end|arrive|return|final|to")

var AttentionIgnoredRe = regexp.MustCompile("attention|attn")
var RegionIgnoredRe = regexp.MustCompile("province|region|other|provincia|bairro|suburb")
var AddressNameIgnoredRe = regexp.MustCompile("address.*nickname|address.*label")
var CompanyRe = regexp.MustCompile("company|business|organization|organisation|firma|firmenname|empresa|societe|société|ragione.?sociale|会社|название.?компании|单位|公司|شرکت|회사|직장")
var AddressLine1Re = regexp.MustCompile("^address$|address[_-]?line(one)?|address1|addr1|street|(?:shipping|billing)address$|strasse|straße|hausnummer|housenumber|house.?name|direccion|dirección|adresse|indirizzo|^住所$|住所1|morada|Адрес|地址|(\\b|_)adres(\\b|_)|^주소.?$|주소.?1")
var AddressLine1LabelRe = regexp.MustCompile("(^\\W*address)|(address\\W*$)|(?:shipping|billing|mailing|pick.?up|drop.?off|delivery|sender|postal|recipient|home|work|office|school|business|mail)[\\s\\-]+address|address\\s+(of|for|to|from)|adresse|indirizzo|住所|地址|(\\b|_)adres(\\b|_) |주소")
var AddressLine2Re = regexp.MustCompile("address[_-]?line(2|two)|address2|addr2|street|suite|unit|adresszusatz|ergänzende.?angaben|direccion2|colonia|adicional|addresssuppl|complementnom|appartement|indirizzo2|住所2|complemento|addrcomplement|Улица|地址2|주소.?2")
var AddressLine2LabelRe = regexp.MustCompile("address2|line2|adresse|indirizzo|地址|주소")
var AddressLinesExtraRe = regexp.MustCompile("address.*line[3-9]|address[3-9]|addr[3-9]|street|line[3-9]|municipio|batiment|residence|indirizzo[3-9]")
var AddressLookupRe = regexp.MustCompile("lookup")
var CountryRe = regexp.MustCompile("country|countries|país|pais|(入国|出国)|国家|국가|나라|(\\b|_)ulce(\\b|_)|کشور")
var CountryLocationRe = regexp.MustCompile("location")
var ZipCodeRe = regexp.MustCompile("zip|postal|post.*code|pcode|pin.?code|postleitzahl|\\bcp\\b|\\bcdp\\b|\\bcap\\b|郵便番号|codigo|codpos|\\bcep\\b|Почтовый.?Индекс|पिन.?कोड|പിന്‍കോഡ്|邮政编码|邮编|郵遞區號|(\\b|_)posta kodu(\\b|_)|우편.?번호")
var Zip4Re = regexp.MustCompile("zip|^-$|post2|codpos2")
var CityRe = regexp.MustCompile("city|town|\\bort\\b|stadt|suburb|ciudad|provincia|localidad|poblacion|ville|commune|localita|市区町村|cidade|Город|市|分區|شهر|शहर|ग्राम|गाँव|നഗരം|ഗ്രാമം|((\\b|_)(il|ilimiz|sehir|kent)(\\b|_))|^시[^도·・]|시[·・]?군[·・]?구")
var StateRe = regexp.MustCompile("state|county|region|province|county|principality|都道府県|estado|provincia|область|省|地區|സംസ്ഥാനം|استان|राज्य|(\\b|_)ilce|ilcemiz(\\b|_)|^시[·・]?도")

/////////////////////////////////////////////////////////////////////////////
// search_field.cc
/////////////////////////////////////////////////////////////////////////////
var SearchTermRe = regexp.MustCompile("^q$|search|query|qry|suche.*|搜索|探す|検索|recherch.*|busca|جستجو|искать|найти|поиск")

/////////////////////////////////////////////////////////////////////////////
// field_price.cc
/////////////////////////////////////////////////////////////////////////////
var PriceRe = regexp.MustCompile("\\bprice\\b|\\brate\\b|\\bcost\\b|قیمة‎|سعر‎|قیمت|\\bprix\\b|\\bcoût\\b|\\bcout\\b|\\btarif\\b")

/////////////////////////////////////////////////////////////////////////////
// credit_card_field.cc
/////////////////////////////////////////////////////////////////////////////
var NameOnCardRe = regexp.MustCompile("card.?(?:holder|owner)|name.*(\\b)?on(\\b)?.*card|(?:card|cc).?name|cc.?full.?name|karteninhaber|nombre.*tarjeta|nom.*carte|nome.*cart|名前|Имя.*карты|信用卡开户名|开户名|持卡人姓名|持卡人姓名")
var NameOnCardContextualRe = regexp.MustCompile("name")
var CardNumberRe = regexp.MustCompile("(add)?(?:card|cc|acct).?(?:number|#|no|num|field)|(telefon|haus|person|fødsels)|nummer|カード番号|Номер.*карты|信用卡号|信用卡号码|信用卡卡號|카드|(numero|número|numéro)|(document|fono|phone|réservation)")
var CardCvcRe = regexp.MustCompile("verification|card.?identification|security.?code|card.?code|security.?value|security.?number|card.?pin|c-v-v|(cvn|cvv|cvc|csc|cvd|cid|ccv)(field)?|\\bcid\\b")

// Expiration date is the most common label here, but some pages have
// Expires, exp. date or exp. month and exp. year.  We also look
// for the field names ccmonth and ccyear, which appear on at least 4 of
// our test pages.

// On at least one page (The China Shop2.html) we find only the labels
// month and year.  So for now we match these words directly) we'll
// see if this turns out to be too general.

// Toolbar Bug 51451: indeed, simply matching month is too general for
//   https://rps.fidelity.com/ftgw/rps/RtlCust/CreatePIN/Init.
// Instead, we match only words beginning with month.
var ExpirationMonthRe = regexp.MustCompile("expir|exp.*mo|exp.*date|ccmonth|cardmonth|addmonth|gueltig|gültig|monat|fecha|date.*exp|scadenza|有効期限|validade|Срок действия карты|月")
var ExpirationYearRe = regexp.MustCompile("exp|^/|(add)?year|ablaufdatum|gueltig|gültig|jahr|fecha|scadenza|有効期限|validade|Срок действия карты|年|有效期")

// Used to match a expiration date field with a two digit year.
// The following conditions must be met:
//  - Exactly two adjacent y's.
//  - (optional) Exactly two adjacent m's before the y's.
//    - (optional) Separated by white-space and/or a dash or slash.
//  - (optional) Prepended with some text similar to Expiration Date.
// Tested in components/autofill/core/common/autofill_regexes_unittest.cc
var ExpirationDate2DigitYearRe = regexp.MustCompile("(?:exp.*date[^y\\n\\r]*|mm\\s*[-/]?\\s*)yy(?:[^y]|$)")

// Used to match a expiration date field with a four digit year.
// Same requirements as|kExpirationDate2DigitYearRe|except:
//  - Exactly four adjacent y's.
// Tested in components/autofill/core/common/autofill_regexes_unittest.cc
var ExpirationDate4DigitYearRe = regexp.MustCompile("(?:exp.*date[^y\\n\\r]*|mm\\s*[-/]?\\s*)yyyy(?:[^y]|$)")

// Used to match expiration date fields that do not specify a year length.
var ExpirationDateRe = regexp.MustCompile("expir|exp.*date|^expfield$|gueltig|gültig|fecha|date.*exp|scadenza|有効期限|validade|Срок действия карты")
var GiftCardRe = regexp.MustCompile("gift.?(card|cert)")
var DebitGiftCardRe = regexp.MustCompile("(?:visa|mastercard|discover|amex|american express).*gift.?card")
var DebitCardRe = regexp.MustCompile("debit.*card")
var DayRe = regexp.MustCompile("day")

/////////////////////////////////////////////////////////////////////////////
// email_field.cc
/////////////////////////////////////////////////////////////////////////////
var EmailRe = regexp.MustCompile("e.?mail|courriel|correo.*electr(o|ó)nico|メールアドレス|Электронной.?Почты|邮件|邮箱|電郵地址|ഇ-മെയില്‍|ഇലക്ട്രോണിക്.? മെയിൽ|ایمیل|پست.*الکترونیک|ईमेल|इलॅक्ट्रॉनिक.?मेल|(\\b|_)eposta(\\b|_)|(?:이메일|전자.?우편|[Ee]-?mail)(.?주소)?")

/////////////////////////////////////////////////////////////////////////////
// name_field.cc
/////////////////////////////////////////////////////////////////////////////
var NameIgnoredRe = regexp.MustCompile("user.?name|user.?id|nickname|maiden name|title|prefix|suffix|vollständiger.?name|用户名|(?:사용자.?)?아이디|사용자.?ID")
var NameRe = regexp.MustCompile("^name|full.?name|your.?name|customer.?name|bill.?name|ship.?name|name.*first.*last|firstandlastname|nombre.*y.*apellidos|^nom(bre)|お名前|氏名|^nome|نام.*نام.*خانوادگی|姓名|(\\b|_)ad soyad(\\b|_)|성명")
var NameSpecificRe = regexp.MustCompile("^name|^nom|^nome")
var FirstNameRe = regexp.MustCompile("first.*name|initials|fname|first$|given.*name|vorname|nombre|forename|prénom|prenom|名|nome|Имя|نام|이름|പേര്|(\\b|_)ad(\\b|_)|नाम")
var MiddleInitialRe = regexp.MustCompile("middle.*initial|m\\.i\\.|mi$|\\bmi\\b")
var MiddleNameRe = regexp.MustCompile("middle.*name|mname|middle$|apellido.?materno|lastlastname")
var LastNameRe = regexp.MustCompile("last.*name|lname|surname|last$|secondname|family.*name|nachname|apellidos?|famille|^nom(bre)|cognome|姓|apelidos|surename|sobrenome|Фамилия|نام.*خانوادگی|उपनाम|മറുപേര്|(\\b|_)soyad(\\b|_)|\\b성(?:[^명]|\\b)")

/////////////////////////////////////////////////////////////////////////////
// phone_field.cc
/////////////////////////////////////////////////////////////////////////////
var PhoneRe = regexp.MustCompile("phone|mobile|contact.?number|telefonnummer|telefono|teléfono|telfixe|電話|telefone|telemovel|телефон|मोबाइल|电话|മൊബൈല്‍|(?:전화|핸드폰|휴대폰|휴대전화)(?:.?번호)?")
var CountryCodeRe = regexp.MustCompile("country.*code|ccode|_cc|phone.*code|user.*phone.*code")
var AreaCodeNotextRe = regexp.MustCompile("^\\($")
var AreaCodeRe = regexp.MustCompile("area.*code|acode|area|지역.?번호")
var PhonePrefixSeparatorRe = regexp.MustCompile("^-$|^\\)$")
var PhoneSuffixSeparatorRe = regexp.MustCompile("^-$")
var PhonePrefixRe = regexp.MustCompile("prefix|exchange|preselection|ddd")
var PhoneSuffixRe = regexp.MustCompile("suffix")
var PhoneExtensionRe = regexp.MustCompile("\\bext|ext\\b|extension|ramal")

/////////////////////////////////////////////////////////////////////////////
// travel_field.cc
/////////////////////////////////////////////////////////////////////////////

var PassportRe = regexp.MustCompile("document.*number|passport|passeport|numero.*documento|pasaporte|書類")
var TravelOriginRe = regexp.MustCompile("point.*of.*entry|arrival|punto.*internaci(o|ó)n|fecha.*llegada|入国")
var TravelDestinationRe = regexp.MustCompile("departure|fecha.*salida|destino|出国")
var FlightRe = regexp.MustCompile("airline|flight|aerol(i|í)nea|n(u|ú)mero.*vuelo|便名|航空会社")

/////////////////////////////////////////////////////////////////////////////
// validation.cc
/////////////////////////////////////////////////////////////////////////////
var UPIVirtualPaymentAddressRe = regexp.MustCompile("^[\\w.+-_]+@(\\w+\\.ifsc\\.npci|aadhaar\\.npci|mobile\\.npci|rupay\\.npci|airtel|airtelpaymentsbank|albk|allahabadbank|allbank|andb|apb|apl|axis|axisbank|axisgo|bandhan|barodampay|birla|boi|cbin|cboi|centralbank|cmsidfc|cnrb|csbcash|csbpay|cub|dbs|dcb|dcbbank|denabank|dlb|eazypay|equitas|ezeepay|fbl|federal|finobank|hdfcbank|hsbc|icici|idbi|idbibank|idfc|idfcbank|idfcnetc|ikwik|imobile|indbank|indianbank|indianbk|indus|iob|jkb|jsb|jsbp|karb|karurvysyabank|kaypay|kbl|kbl052|kmb|kmbl|kotak|kvb|kvbank|lime|lvb|lvbank|mahb|obc|okaxis|okbizaxis|okhdfcbank|okicici|oksbi|paytm|payzapp|pingpay|pnb|pockets|psb|purz|rajgovhdfcbank|rbl|sbi|sc|scb|scbl|scmobile|sib|srcb|synd|syndbank|syndicate|tjsb|tjsp|ubi|uboi|uco|unionbank|unionbankofindia|united|upi|utbi|vijayabank|vijb|vjb|ybl|yesbank|yesbankltd)$")

var InternationalBankAccountNumberRe = regexp.MustCompile("^[a-zA-Z]{2}[0-9]{2}[a-zA-Z0-9]{4}[0-9]{7}([a-zA-Z0-9]?){0,16}$")

// Matches all 3 and 4 digit numbers.
var CreditCardCVCPattern = regexp.MustCompile("^\\d{3,4}$")

// Matches numbers in the range [2010-2099].
var CreditCard4DigitExpYearPattern = regexp.MustCompile("^[2][0][1-9][0-9]$")

/////////////////////////////////////////////////////////////////////////////
// form_structure.cc
/////////////////////////////////////////////////////////////////////////////
var UrlSearchActionRe = regexp.MustCompile("/search(/|((\\w*\\.\\w+)?$))")

/////////////////////////////////////////////////////////////////////////////
// form_parser.cc
/////////////////////////////////////////////////////////////////////////////
var SocialSecurityRe = regexp.MustCompile("ssn|social.?security.?(num(ber)?|#)*")
var PasswordRe = regexp.MustCompile("password|passwd|pwd|pass|パスワード")
var OneTimePwdRe = regexp.MustCompile("one.?time|sms.?(code|token|password|pwd|pass)")

// Others
var CommentTitleRe = regexp.MustCompile("title|sub(ject)")
var CommentRe = regexp.MustCompile("comment|reply|answer|message|text|msg|resp(ponse)")

var NetworkMaskRe = regexp.MustCompile("netmask|network_mask|subnet(_mask)")
var IPAddressRe = regexp.MustCompile("ip|ip(_addr)|ip_address|gateway|ipv4.+")
var IPV6AddressRe = regexp.MustCompile("ipv6|ipv6(_addr)|ipv6_address|ipv6.+")
