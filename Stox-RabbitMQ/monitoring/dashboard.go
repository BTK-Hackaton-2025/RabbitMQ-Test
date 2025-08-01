package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ServiceStatus represents the status of a microservice
type ServiceStatus struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Uptime     string    `json:"uptime"`
	CPU        string    `json:"cpu"`
	Memory     string    `json:"memory"`
	LastSeen   time.Time `json:"last_seen"`
	Health     string    `json:"health"`
	Containers int       `json:"containers"`
}

// QueueInfo represents RabbitMQ queue information
type QueueInfo struct {
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
	Consumers int    `json:"consumers"`
	State     string `json:"state"`
}

// SystemStats represents overall system statistics
type SystemStats struct {
	Services      []ServiceStatus `json:"services"`
	Queues        []QueueInfo     `json:"queues"`
	TotalMessages int             `json:"total_messages"`
	ActiveWorkers int             `json:"active_workers"`
	Uptime        time.Duration   `json:"uptime"`
	LastUpdate    time.Time       `json:"last_update"`
}

var startTime = time.Now()

func main() {
	log.Println("🎛️  Starting Stox Monitoring Dashboard...")

	// Serve static files
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/api/status", statusAPIHandler)
	http.HandleFunc("/api/services", servicesAPIHandler)
	http.HandleFunc("/api/queues", queuesAPIHandler)
	http.HandleFunc("/api/restart", restartServiceHandler)
	http.HandleFunc("/api/scale", scaleServiceHandler)

	log.Println("📊 Dashboard available at: http://localhost:8080")
	log.Println("🔧 API endpoints:")
	log.Println("   GET  /api/status   - System status")
	log.Println("   GET  /api/services - Service details")
	log.Println("   GET  /api/queues   - Queue information")
	log.Println("   POST /api/restart  - Restart service")
	log.Println("   POST /api/scale    - Scale service")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stox RabbitMQ Monitoring Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f6fa; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; }
        .header h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        .header p { opacity: 0.9; }
        .container { max-width: 1400px; margin: 0 auto; padding: 2rem; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
        .card { background: white; border-radius: 10px; padding: 1.5rem; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .card h3 { color: #2c3e50; margin-bottom: 1rem; border-bottom: 2px solid #ecf0f1; padding-bottom: 0.5rem; }
        .metric { display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 0; border-bottom: 1px solid #ecf0f1; }
        .metric:last-child { border-bottom: none; }
        .metric-label { font-weight: 500; color: #34495e; }
        .metric-value { font-weight: bold; }
        .status-running { color: #27ae60; }
        .status-stopped { color: #e74c3c; }
        .status-warning { color: #f39c12; }
        .btn { background: #3498db; color: white; border: none; padding: 0.5rem 1rem; border-radius: 5px; cursor: pointer; margin: 0.25rem; }
        .btn:hover { background: #2980b9; }
        .btn-danger { background: #e74c3c; }
        .btn-danger:hover { background: #c0392b; }
        .btn-success { background: #27ae60; }
        .btn-success:hover { background: #229954; }
        .refresh-btn { position: fixed; bottom: 2rem; right: 2rem; background: #667eea; border: none; color: white; padding: 1rem; border-radius: 50%; cursor: pointer; box-shadow: 0 4px 15px rgba(0,0,0,0.2); }
        .refresh-btn:hover { background: #764ba2; }
        .queue-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; }
        .queue-item { background: #f8f9fa; padding: 1rem; border-radius: 8px; border-left: 4px solid #3498db; }
        .service-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem; }
        .service-item { background: #f8f9fa; padding: 1rem; border-radius: 8px; }
        .service-controls { margin-top: 1rem; }
        .log-viewer { background: #2c3e50; color: #ecf0f1; padding: 1rem; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 0.9rem; max-height: 300px; overflow-y: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎛️ Stox RabbitMQ Monitoring Dashboard</h1>
        <p>Real-time monitoring and management for your e-commerce automation platform</p>
    </div>

    <div class="container">
        <!-- System Overview -->
        <div class="grid">
            <div class="card">
                <h3>📊 System Overview</h3>
                <div class="metric">
                    <span class="metric-label">Platform Uptime</span>
                    <span class="metric-value" id="uptime">Loading...</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Active Services</span>
                    <span class="metric-value" id="active-services">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Total Messages</span>
                    <span class="metric-value" id="total-messages">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Active Workers</span>
                    <span class="metric-value" id="active-workers">-</span>
                </div>
            </div>

            <div class="card">
                <h3>🔧 Quick Actions</h3>
                <button class="btn" onclick="restartAllServices()">Restart All Services</button>
                <button class="btn btn-success" onclick="scaleAIService()">Scale AI Workers</button>
                <button class="btn" onclick="showLogs()">View Logs</button>
                <button class="btn btn-danger" onclick="emergencyStop()">Emergency Stop</button>
            </div>

            <div class="card">
                <h3>🐰 RabbitMQ Status</h3>
                <div class="metric">
                    <span class="metric-label">Management UI</span>
                    <span class="metric-value">
                        <a href="http://localhost:15672" target="_blank" style="color: #3498db;">Open Dashboard</a>
                    </span>
                </div>
                <div class="metric">
                    <span class="metric-label">Connection</span>
                    <span class="metric-value status-running" id="rabbitmq-status">Connected</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Total Queues</span>
                    <span class="metric-value" id="total-queues">-</span>
                </div>
            </div>
        </div>

        <!-- Services Status -->
        <div class="card">
            <h3>🚀 Microservices Status</h3>
            <div class="service-grid" id="services-grid">
                <!-- Services will be loaded here -->
            </div>
        </div>

        <!-- Queue Status -->
        <div class="card">
            <h3>📬 Message Queues</h3>
            <div class="queue-grid" id="queues-grid">
                <!-- Queues will be loaded here -->
            </div>
        </div>

        <!-- System Logs -->
        <div class="card">
            <h3>📝 Recent Activity</h3>
            <div class="log-viewer" id="log-viewer">
                Loading system logs...
            </div>
        </div>
    </div>

    <button class="refresh-btn" onclick="refreshData()" title="Refresh Data">
        🔄
    </button>

    <script>
        let refreshInterval;

        function refreshData() {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    updateSystemOverview(data);
                    updateServices(data.services);
                    updateQueues(data.queues);
                })
                .catch(error => console.error('Error fetching data:', error));
        }

        function updateSystemOverview(data) {
            document.getElementById('uptime').textContent = formatDuration(data.uptime);
            document.getElementById('active-services').textContent = data.services.filter(s => s.status === 'running').length + '/' + data.services.length;
            document.getElementById('total-messages').textContent = data.total_messages.toLocaleString();
            document.getElementById('active-workers').textContent = data.active_workers;
            document.getElementById('total-queues').textContent = data.queues.length;
        }

        function updateServices(services) {
            const grid = document.getElementById('services-grid');
            grid.innerHTML = services.map(service => ` + "`" + `
                <div class="service-item">
                    <div class="metric">
                        <span class="metric-label">${ + "`" + `service.name}</span>
                        <span class="metric-value status-${ + "`" + `service.status === 'running' ? 'running' : 'stopped'}">
                            ${ + "`" + `service.status}
                        </span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Containers</span>
                        <span class="metric-value">${ + "`" + `service.containers}</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Memory</span>
                        <span class="metric-value">${ + "`" + `service.memory}</span>
                    </div>
                    <div class="service-controls">
                        <button class="btn" onclick="restartService('${ + "`" + `service.name}')">Restart</button>
                        <button class="btn" onclick="viewServiceLogs('${ + "`" + `service.name}')">Logs</button>
                        <button class="btn btn-success" onclick="scaleService('${ + "`" + `service.name}')">Scale</button>
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        function updateQueues(queues) {
            const grid = document.getElementById('queues-grid');
            grid.innerHTML = queues.map(queue => `
                <div class="queue-item">
                    <div class="metric">
                        <span class="metric-label">${queue.name}</span>
                        <span class="metric-value">${queue.messages} msgs</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Consumers</span>
                        <span class="metric-value">${queue.consumers}</span>
                    </div>
                </div>
            `).join('');
        }

        function formatDuration(nanoseconds) {
            const seconds = Math.floor(nanoseconds / 1000000000);
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = seconds % 60;
            return hours > 0 ? ${hours}h ${minutes}m ${secs}s : ${minutes}m ${secs}s;
        }

        function restartService(serviceName) {
            if (confirm('Restart ' + serviceName + '?')) {
                fetch('/api/restart', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ service: serviceName })
                });
            }
        }

        function scaleService(serviceName) {
            const replicas = prompt('Number of replicas for ' + serviceName + ':', '1');
            if (replicas) {
                fetch('/api/scale', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ service: serviceName, replicas: parseInt(replicas) })
                });
            }
        }

        function restartAllServices() {
            if (confirm('Restart all services? This will cause temporary downtime.')) {
                fetch('/api/restart', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ service: 'all' })
                });
            }
        }

        function scaleAIService() {
            const replicas = prompt('Number of AI workers:', '3');
            if (replicas) {
                fetch('/api/scale', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ service: 'ai-service', replicas: parseInt(replicas) })
                });
            }
        }

        function emergencyStop() {
            if (confirm('EMERGENCY STOP: This will stop all services immediately!')) {
                fetch('/api/restart', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ service: 'stop-all' })
                });
            }
        }

        function showLogs() {
            window.open('/api/logs', '_blank');
        }

        // Initialize dashboard
        document.addEventListener('DOMContentLoaded', function() {
            refreshData();
            refreshInterval = setInterval(refreshData, 5000); // Refresh every 5 seconds
        });
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("dashboard").Parse(tmpl))
	t.Execute(w, nil)
}

func statusAPIHandler(w http.ResponseWriter, r *http.Request) {
	services := getServiceStatus()
	queues := getQueueInfo()
	
	stats := SystemStats{
		Services:      services,
		Queues:        queues,
		TotalMessages: getTotalMessages(queues),
		ActiveWorkers: getActiveWorkers(services),
		Uptime:        time.Since(startTime),
		LastUpdate:    time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func servicesAPIHandler(w http.ResponseWriter, r *http.Request) {
	services := getServiceStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func queuesAPIHandler(w http.ResponseWriter, r *http.Request) {
	queues := getQueueInfo()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queues)
}

func restartServiceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string `json:"service"`
	}
	
	json.NewDecoder(r.Body).Decode(&req)
	
	var cmd *exec.Cmd
	if req.Service == "all" {
		cmd = exec.Command("./docker-manager.sh", "restart")
	} else if req.Service == "stop-all" {
		cmd = exec.Command("./docker-manager.sh", "stop")
	} else {
		cmd = exec.Command("docker-compose", "-p", "stox", "restart", req.Service)
	}
	
	output, err := cmd.CombinedOutput()
	
	response := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func scaleServiceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service  string `json:"service"`
		Replicas int    `json:"replicas"`
	}
	
	json.NewDecoder(r.Body).Decode(&req)
	
	cmd := exec.Command("./docker-manager.sh", "scale", req.Service, fmt.Sprintf("%d", req.Replicas))
	output, err := cmd.CombinedOutput()
	
	response := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getServiceStatus() []ServiceStatus {
	cmd := exec.Command("docker-compose", "-p", "stox", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return []ServiceStatus{}
	}
	
	services := []ServiceStatus{}
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var container map[string]interface{}
		if json.Unmarshal([]byte(line), &container) == nil {
			status := "stopped"
			if state, ok := container["State"].(string); ok && state == "running" {
				status = "running"
			}
			
			service := ServiceStatus{
				Name:       fmt.Sprintf("%v", container["Service"]),
				Status:     status,
				Containers: 1,
				LastSeen:   time.Now(),
				Health:     "healthy",
			}
			services = append(services, service)
		}
	}
	
	return services
}

func getQueueInfo() []QueueInfo {
	cmd := exec.Command("docker", "exec", "stox-rabbitmq", "rabbitmqctl", "list_queues", "name", "messages", "consumers", "--formatter", "json")
	output, err := cmd.Output()
	if err != nil {
		return []QueueInfo{}
	}
	
	var result [][]interface{}
	if json.Unmarshal(output, &result) != nil {
		return []QueueInfo{}
	}
	
	queues := []QueueInfo{}
	for _, item := range result {
		if len(item) >= 3 {
			queue := QueueInfo{
				Name:      fmt.Sprintf("%v", item[0]),
				Messages:  int(item[1].(float64)),
				Consumers: int(item[2].(float64)),
				State:     "running",
			}
			queues = append(queues, queue)
		}
	}
	
	return queues
}

func getTotalMessages(queues []QueueInfo) int {
	total := 0
	for _, queue := range queues {
		total += queue.Messages
	}
	return total
}

func getActiveWorkers(services []ServiceStatus) int {
	workers := 0
	for _, service := range services {
		if service.Status == "running" && strings.Contains(service.Name, "service") {
			workers += service.Containers
		}
	}
	return workers
}
