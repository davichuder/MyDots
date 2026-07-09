# MyDots en WSL2

MyDots no se ejecuta directamente en Windows. Necesitás WSL2 (Windows Subsystem for Linux) con una distribución Ubuntu.

## Prerequisitos

- Windows 10 version 2004+ o Windows 11
- Virtualization habilitada en BIOS/UEFI

## Instalación de WSL2

1. Abrí PowerShell como Administrador y ejecutá:

   ```powershell
   wsl --install
   ```

2. Reiniciá la máquina cuando termine.

3. Después del reinicio, Ubuntu se abre automáticamente. Creá tu usuario y contraseña de Linux.

4. Verificá que tenés WSL2:

   ```powershell
   wsl -l -v
   ```

   La columna "Version" debe mostrar 2.

## Cómo ejecutar MyDots

Dentro de WSL2 (desde Ubuntu):

```bash
git clone https://github.com/davichuder/MyDots
cd MyDots
go run .
```

O via Homebrew:

```bash
brew tap davichuder/homebrew-tap
brew install mydots
mydots
```

## Notas

- WSL2 corre Linux nativamente. Todo lo que instala MyDots funciona dentro de WSL2.
- Los programas con GUI (como Ghostty) requieren WSLg, que viene incluido en versiones recientes de Windows.
- Las fuentes Nerd Font se instalan dentro del entorno Linux y están disponibles para apps WSLg, pero NO para la terminal de Windows host.
- Para acceder a tus archivos de Windows: `/mnt/c/`

Presioná cualquier tecla para salir.
