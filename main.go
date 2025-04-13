package main

import (
	"log"
	"os"
	"os/exec"
	"time"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/tebeka/selenium"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)
var chromeCmd *exec.Cmd

func main() {
	defer func() {
		if r := recover(); r != nil {
			logger.Fatalf("❌ Application crashed: %v", r)
		}
	}()

	logger.Println("🚀 Resume Auto-Updater starting up...")

	// Load env variables
	if err := godotenv.Load(); err != nil {
		logger.Fatalf("❌ Failed to load .env: %v", err)
	}
	username := os.Getenv("NAUKRI_USERNAME")
	password := os.Getenv("NAUKRI_PASSWORD")
	if username == "" || password == "" {
		logger.Fatal("❌ NAUKRI_USERNAME or NAUKRI_PASSWORD missing in .env")
	}

	// Start ChromeDriver
	startChromeDriver()
	defer stopChromeDriver()

	// Handle graceful shutdown
	setupGracefulShutdown()

	// Wait for ChromeDriver to be ready
	waitForChromeDriver()

	// Schedule jobs
	c := cron.New()
	logger.Println("📅 Starting scheduler...")

	go updateResume(username, password) // Run immediately on start

	c.AddFunc("0 9 * * *", func() { updateResume(username, password) })
	c.AddFunc("0 14 * * *", func() { updateResume(username, password) })
	c.AddFunc("0 18 * * *", func() { updateResume(username, password) })

	c.Start()
	logger.Println("✅ Scheduler started. Waiting for job triggers...")

	select {} // Keep running
}

func startChromeDriver() {
	logger.Println("🚀 Starting ChromeDriver...")
	chromeCmd = exec.Command("chromedriver", "--port=9515")
	chromeCmd.Stdout = os.Stdout
	chromeCmd.Stderr = os.Stderr
	err := chromeCmd.Start()
	if err != nil {
		logger.Fatalf("❌ Failed to start ChromeDriver: %v", err)
	}
}

func waitForChromeDriver() {
	logger.Println("⏳ Waiting for ChromeDriver to become ready...")
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://localhost:9515/status")
		if err == nil && resp.StatusCode == http.StatusOK {
			logger.Println("✅ ChromeDriver is ready.")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	logger.Fatalf("❌ ChromeDriver did not become ready in time.")
}

func stopChromeDriver() {
	if chromeCmd != nil && chromeCmd.Process != nil {
		_ = chromeCmd.Process.Kill()
		logger.Println("🛑 ChromeDriver stopped.")
	}
}

func setupGracefulShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		logger.Println("⚠️ Received shutdown signal.")
		stopChromeDriver()
		os.Exit(0)
	}()
}

func updateResume(username, password string) {
	logger.Println("▶ Starting resume update process...")

	caps := selenium.Capabilities{
		"browserName": "chrome",
		"goog:chromeOptions": map[string]interface{}{
			"args": []string{"--headless", "--disable-gpu", "--no-sandbox"},
		},
	}

	wd, err := selenium.NewRemote(caps, "http://localhost:9515/wd/hub")
	if err != nil {
		logger.Printf("❌ Failed to start Selenium session: %v", err)
		return
	}
	defer wd.Quit()

	// Open login page
	err = wd.Get("https://www.naukri.com/mnjuser/profile")
	if err != nil {
		logger.Printf("❌ Failed to load Naukri profile page: %v", err)
		return
	}

	// Fill login
	usernameField, err := wd.FindElement(selenium.ByID, "usernameField")
	if err != nil {
		logger.Printf("❌ Username field not found: %v", err)
		return
	}
	passwordField, err := wd.FindElement(selenium.ByID, "passwordField")
	if err != nil {
		logger.Printf("❌ Password field not found: %v", err)
		return
	}
	loginBtn, err := wd.FindElement(selenium.ByXPATH, "//button[@type='submit']")
	if err != nil {
		logger.Printf("❌ Login button not found: %v", err)
		return
	}
	usernameField.SendKeys(username)
	passwordField.SendKeys(password)
	loginBtn.Click()

	time.Sleep(5 * time.Second) // Wait for login

	// Edit resume headline
	updateBtn, err := wd.FindElement(selenium.ByXPATH, "//span[text()='Resume headline']/following::span[text()='edit']")
	if err != nil {
		logger.Printf("❌ Resume headline edit not found: %v", err)
		return
	}
	updateBtn.Click()
	time.Sleep(2 * time.Second)

	saveBtn, err := wd.FindElement(selenium.ByXPATH, "//button[text()='Save']")
	if err != nil {
		logger.Printf("❌ Save button not found: %v", err)
		return
	}
	saveBtn.Click()

	logger.Println("✅ Resume updated successfully.")
}
