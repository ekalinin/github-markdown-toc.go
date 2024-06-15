/*
Слои:
- приложение - собирает всё вместе
- контролер - обработчики внешних систем
- сущности
- useсase - основная БЛ
  - локальный md-файл
  - удалённый md-файл
  - удалённый html-файл
*/
package app

import (
	"os"

	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

func (a *App) Run() error {

	// do not show for stdin case (Files is empty)
	if !a.cfg.HideHeader && len(a.cfg.Files) == 1 {
		utils.ShowHeader(os.Stdout)
	}

	if err := a.ctl.Process(); err != nil {
		return err
	}

	if !a.cfg.HideFooter {
		utils.ShowFooter(os.Stdout)
	}

	return nil
}
