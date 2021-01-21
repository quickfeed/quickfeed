# Guide to setting up the Windows Subsystem for Linux

## Background

For development on Windows machines we recommend using the Windows Subsystem for Linux (WSL).
This allows you to run a lightweight Linux virtual machine on your Windows system where you can use the shell commands and utilities presented in the course.

There are 2 versions of WSL - WSL1 and WSL2.
WSL2 is advertised to be much faster and supports more functions of the Linux kernel, but it is more isolated from your file system.
It also requires a recent version of Windows 10 (version 2004, build 19041 or higher).
We suggest using WSL2 for its performance improvements and other capabilities.
You can read more about the differences between the WSL versions [here](https://docs.microsoft.com/en-us/windows/wsl/compare-versions).

## Installation

Follow [this tutorial from Microsoft](https://docs.microsoft.com/en-us/windows/wsl/install-win10) to set up WSL(2).

### Installing WSL

In short you need to run PowerShell as the administrator and issue the following commands:

Install WSL:

```powershell
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
```

If you simply wish to use WSL1, you can restart your computer at this point.

### Updating to WSL2

You need to enable the "Virtual Machine Platform" component to update to WSL2.

```powershell
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart
```

You will need to restart for the changes to take effect.

Finally, set WSL2 as the default WSL version.

```powershell
wsl --set-default-version 2
```

You might be prompted to update the kernel with a message such as this:

```powershell
> wsl --set-default-version 2
WSL 2 requires an update to its kernel component. For information please visit https://aka.ms/wsl2kernel
```

In that case, follow the instructions at the [link](https://aka.ms/wsl2kernel) to update the kernel component.

### Installing a Linux distribution

After these steps you can install your Linux distribution of choice from the Microsoft Store.
We recommend Ubuntu 20.04 LTS since there are a lot of resources for beginners online.
After installing your Linux distribution of choice you can launch it as a regular application.
Further we recommend installing and using the more feature rich [Windows Terminal](https://www.microsoft.com/en-us/p/windows-terminal/9n0dx20hk701) over the default terminal.
We also provide a [Go installation guide](setup-go.md) and a some guidance related to [code editors](setup-editors.md).

## Confirming WSL2 is being used

To confirm that WSL2 is being used, run the following command in PowerShell:

```powershell
wsl --list --verbose
```

If `VERSION` is not `2` you need to manually set the version.
E.g. if your WSL distribution is called `Ubuntu-20.04` you can execute the following command:

```powershell
wsl --set-version Ubuntu-20.04 2
```

## Troubleshooting

If you run into issues you might want to check the [WSL issue tracker](https://github.com/Microsoft/WSL/issues).

If there are some blocking issues, you might want to switch back to WSL1 for the meantime.
In PowerShell, you should be able to switch to WSL1 by finding your WSL distro's name (let's call it `<distro>`) with:

```powershell
wsl --list --verbose
```

and then changing the version of `<distro>` with:

```powershell
wsl --set-version <distro> 1
```
