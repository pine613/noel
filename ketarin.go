package main

import (
    "fmt"
    "os"
    "os/exec"
    "io"
    "io/ioutil"
    "path/filepath"
    "time"
    "errors"
    "encoding/xml"
)

var KetarinAppDataDirName = "Ketarin"
var DatabaseFileName = "jobs.db"
var ChocopkgupConfigPath = `C:\tools\ChocolateyPackageUpdater\chocopkgup.exe.config`
var SettingFileName = "Ketarin.xml"

type AppSettings struct {
    XMLName xml.Name `xml:"appSettings"`
    Adds    []*Add   `xml:"add"`
}

type Add struct {
    XMLName xml.Name `xml:"add"`
    Key     string   `xml:"key,attr"`
    Value   string   `xml:"value,attr"`
}

type SupportedRuntime struct {
    XMLName xml.Name `xml:"supportedRuntime"`
    Version string   `xml:"version,attr"`
    Sku     string   `xml:"sku,attr"`
}

type Startup struct {
    XMLName           xml.Name          `xml:"startup"`
    SupportedRuntime  *SupportedRuntime
}

type Configuration struct {
    XMLName     xml.Name      `xml:"configuration"`
    AppSettings *AppSettings  `xml:"appSettings"`
    Startup     *Startup      `xml:"startup"`
}

func getKetarinDatabase() string {
    appdata := filepath.Join(os.Getenv("APPDATA"), KetarinAppDataDirName)
    dbPath := filepath.Join(appdata, DatabaseFileName)
    
    return dbPath
}

func getKetarinDatabaseBackup() string {
    appdata := filepath.Join(os.Getenv("APPDATA"), KetarinAppDataDirName)
    date := time.Now().Format("2006-01-02-150405")
    dbPath := filepath.Join(appdata, DatabaseFileName + "_" + date + ".noel.bak")
    
    return dbPath
}

func SwapKetarinDatabase() error {
    dbFile, err := os.Open(getKetarinDatabase())
    
    if err != nil {
        return err
    }
    
    defer dbFile.Close()
    
    destFile, err := os.Create(DatabaseFileName)
    
    if err != nil {
        return err
    }
    
    defer destFile.Close()
    
    
    bakPath := getKetarinDatabaseBackup()
    bakFile, err := os.Create(bakPath)
    
    if err != nil {
        return err
    }
    
    defer bakFile.Close()
    
    if _, err = io.Copy(destFile, dbFile); err != nil {
        return err
    }
    
    _, err = io.Copy(bakFile, dbFile)
    
    return err
}

func RestoreKetarinDatabase() error {
    dbFile, err := os.Create(getKetarinDatabase())
    
    if err != nil {
        return err
    }
    
    defer dbFile.Close()
    
    swapFile, err := os.Open(DatabaseFileName)
    
    if err != nil {
        return err
    }
    
    defer swapFile.Close()
    
    _, err = io.Copy(dbFile, swapFile)
    return err
}

func ClearKetarinDatabase() error {
    dbPath := getKetarinDatabase()
    return os.Remove(dbPath)
}


func SetChocopkgupPackageFolder(pkgDir string) error {
    fmt.Println("> Change chocopkgup.exe.config")
    
    path := ChocopkgupConfigPath
    xmlFile, err := ioutil.ReadFile(path)
    
    if err != nil {
        return err
    }
    
    var conf Configuration
    if err = xml.Unmarshal(xmlFile, &conf); err != nil {
        return err
    }
    
    appSettings := conf.AppSettings
    
    for _, add := range(appSettings.Adds) {
        if add.Key == "PackagesFolder" {
            // already same directory
            if add.Value == pkgDir {
                return nil
            }
            
            fmt.Println("PackagesFolder (old): " + add.Value)
            fmt.Println("PackagesFolder (new): " + pkgDir)
            
            add.Value = pkgDir
            newXmlFile, err := xml.Marshal(conf)
            
            if err != nil {
                return err
            }
            
            validXmlFile := xml.Header + string(newXmlFile)
            err = ioutil.WriteFile(path, []byte(validXmlFile), os.ModePerm)
            if err != nil {
                return err
            } else {
                return nil
            }
        }
    }
    
    return errors.New(`Can't find PackagesFolder setting`)
}

func WaitKetarinProcess() error {
    cmd := exec.Command(
        "powershell",
        "-NoProfile", "-ExecutionPolicy", "unrestricted",
        "-Command", "Wait-Process -Name ketarin")
    
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    return cmd.Run()
}

func RunKetarin() error {
    tmpDir, err := ioutil.TempDir("", "nodel")
    
    if err != nil {
        return err
    }
    
    logPath := filepath.Join(tmpDir, "ketarin.log")
    cmd := exec.Command("ketarin", "/silent", "/notify", "/log=" + logPath)
    
    if err := cmd.Run(); err != nil {
        return err
    }
    
    if err:= WaitKetarinProcess(); err != nil {
        return err
    }
    
    log, err := ioutil.ReadFile(logPath);
    
    if err != nil {
        return err
    }
    
    fmt.Println(string(log))
    
    return os.RemoveAll(tmpDir)
}

func InstallKetarinSetting(data TestData) error {
    settingPath := filepath.Join(data.Name, SettingFileName)
    
    if _, err := os.Stat(settingPath); err != nil {
        return errors.New("Setting file not found!\n" + settingPath)
    }
    
    cmd := exec.Command("ketarin", "/install=" + settingPath, "/exit")
    
    if err := cmd.Run(); err != nil {
        return err
    }
    
    return WaitKetarinProcess()
}