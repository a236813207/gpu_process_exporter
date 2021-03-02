package main

import (
    "fmt"
    "net/http"
    "log"
    "os"
    "os/exec"
    "encoding/xml"
    "strings"
    "strconv"
)

// 解析XML,定义结构体
type Processes struct {  
    XMLName xml.Name `xml:"processes"`
    ProcessInfo []ProcessInfo `xml:"process_info"`
}

type ProcessInfo struct {  
    XMLName xml.Name `xml:"process_info"`
    Pid string `xml:"pid"`  
    Type string `xml:"type"`
    ProcessName string `xml:"process_name"`  
    UsedMemory string `xml:"used_memory"`
}

type Pi struct {
    Value string `xml:",innerxml"`
}

type Pros struct {
    Pi []Pi `xml:"gpu>processes"`
}

type FbMemoryUsage struct {
    Total []string `xml:"gpu>fb_memory_usage>total"`
    Used []string `xml:"gpu>fb_memory_usage>used"`
    Free []string `xml:"gpu>fb_memory_usage>free"`
}


// Pid, Type, ProcessName, UsedMemory
func metrics(response http.ResponseWriter, request *http.Request) {
    out, err := exec.Command(
        "nvidia-smi",
        "-q",
        "-x").Output()

    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }
    

    /*var out = `
        <?xml version="1.0" encoding="utf-8"?>
        <nvidia_smi_log>
            <product_name>Tesla P4</product_name>
            <gpu id="00000000:02.0">
                <fb_memory_usage>
                    <total>7606 MiB</total>
                    <used>2141 MiB</used>
                    <free>5465 MiB</free>
                </fb_memory_usage>
                <processes>
                    <process_info>
                        <pid>408</pid>
                        <type>C</type>
                        <process_name>/home/miniconda3/bin/python</process_name>
                        <used_memory>1849 MiB</used_memory>
                    </process_info>
                    <process_info>
                        <pid>3049</pid>
                        <type>C</type>
                        <process_name>nginx</process_name>
                        <used_memory>135 MiB</used_memory>
                    </process_info>
                </processes>
                <accounted_processes>
                </accounted_processes>
            </gpu>

            <gpu id="00000000:03.0">
                <fb_memory_usage>
                    <total>7606 MiB</total>
                    <used>3141 MiB</used>
                    <free>4465 MiB</free>
                </fb_memory_usage>
                <processes>
                    <process_info>
                        <pid>4081</pid>
                        <type>C</type>
                        <process_name>/home/miniconda3/bin/python</process_name>
                        <used_memory>1849 MiB</used_memory>
                    </process_info>
                    <process_info>
                        <pid>30491</pid>
                        <type>C</type>
                        <process_name>nginx</process_name>
                        <used_memory>135 MiB</used_memory>
                    </process_info>
                </processes>
                <accounted_processes>
                </accounted_processes>
            </gpu>
        </nvidia_smi_log>
    `*/
    pros := Pros{[]Pi{}}
    //反序列化xml
    xml.Unmarshal(out, &pros)
    //xml.Unmarshal([]byte(out), &pros)
    //fmt.Printf("pros:%s\n", pros)
    result := ""
    for index, pi := range pros.Pi {
        var prosXml = fmt.Sprintf("%s%s%s","<processes>\n",pi.Value,"\n</processes>")
        //fmt.Printf("prosXml:%s\n", prosXml)

        ps := Processes{}
        xml.Unmarshal([]byte(prosXml), &ps)

        for _, info := range ps.ProcessInfo {
            //fmt.Printf("%s\n", info)
            intUserdMemory,err := strconv.Atoi(strings.Fields(info.UsedMemory)[0])
            if err != nil {
                intUserdMemory=-1
            }
            result += fmt.Sprintf("GPU_EXPORTER{GPU=\"GPU%d\",Pid=\"%s\",Type=\"%s\",ProcessName=\"%s\"} %d\n", index, info.Pid, info.Type, info.ProcessName, intUserdMemory)
        }
    }
    

    fmu := FbMemoryUsage{}
    xml.Unmarshal([]byte(out), &fmu)
    fmt.Printf("fmu: %s\n", fmu)

    if fmu.Total != nil && len(fmu.Total)>0 {
        for index, total := range fmu.Total {
            intTotal,err := strconv.Atoi(strings.Fields(total)[0])
            if err != nil {
                intTotal=-1
            }
            result += fmt.Sprintf("GPU_EXPORTER{GPU=\"GPU%d\",Pid=\"%s\",Type=\"%s\",ProcessName=\"%s\"} %d\n", index, "", "", "Total", intTotal)
        }
    }
    
    if fmu.Used != nil && len(fmu.Used)>0 {
        for index, used := range fmu.Used {
            intUsed,err := strconv.Atoi(strings.Fields(used)[0])
            if err != nil {
                intUsed=-1
            }
            result += fmt.Sprintf("GPU_EXPORTER{GPU=\"GPU%d\",Pid=\"%s\",Type=\"%s\",ProcessName=\"%s\"} %d\n", index, "", "", "Used", intUsed)
        }
    }

    if fmu.Free != nil && len(fmu.Free)>0 {
        for index, free := range fmu.Free {
            intFree,err := strconv.Atoi(strings.Fields(free)[0])
            if err != nil {
                intFree=-1
            }
            result += fmt.Sprintf("GPU_EXPORTER{GPU=\"GPU%d\",Pid=\"%s\",Type=\"%s\",ProcessName=\"%s\"} %d\n", index, "", "", "Free", intFree)
        }
    }

    fmt.Printf("result:%s\n", result)
    fmt.Fprintf(response, result)
}

func main() {
    addr := ":9102"
    if len(os.Args) > 1 {
        addr = ":" + os.Args[1]
    }

    http.HandleFunc("/metrics", metrics)
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
