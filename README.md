# MGDialogOpenFile

> [!NOTE]
> This project is no longer maintained, I recommend using the [MGDialogBox](https://github.com/mugomes/mgdialogbox).

MGDialogOpenFile é um componente sofisticado para Fyne que abre uma caixa de dialogo para abrir arquivo ou múltiplos arquivos.

## Recursos

- Filtros por tipo de arquivo
- Campo de Pesquisa
- Salva o diretório mais recente aberto
- Janela Redimensionavel
- Suporte para Múltiplos Arquivos

## Instalação

`go get github.com/mugomes/mgdialogopenfile`

## Exemplo

```
import "github.com/mugomes/mgdialogopenfile"

fs := mgdialogopenfile.Show(a, "Abrir Arquivo", []string{".webp", ".jpg", ".png"}, true, func(filenames []string) {
	for _, filename := range filenames {
        print(filename)
    }
})
```

## Information

 - [Page MGDialogOpenFile](https://github.com/mugomes/mgdialogopenfile)

## Requirement

 - Go 1.25.3
 - Fyne 2.7.0

## Support

- GitHub: https://github.com/sponsors/mugomes
- More: https://mugomes.github.io/apoie.html

## License

Copyright (c) 2025 Murilo Gomes Julio

Licensed under the [MIT](https://github.com/mugomes/mgdialogopenfile/blob/main/LICENSE) license.

All contributions to the MGDialogOpenFile are subject to this license.
