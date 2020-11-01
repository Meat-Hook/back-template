package generated

//go:generate rm -rf models restapi client
//go:generate swagger generate server -f ../../../../swagger.yml --strict-responders --strict-additional-properties --keep-spec-order --principal github.com/Meat-Hook/back-template/internal/modules/session/internal/app.Session --exclude-main
//go:generate swagger generate client -f ../../../../swagger.yml
