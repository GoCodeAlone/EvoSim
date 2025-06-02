#!/bin/bash

# Demonstration script for the Statistical Analysis System
echo "🧬 EvoSim Statistical Analysis System Demo"
echo "=========================================="
echo ""

echo "Building the simulation..."
GOWORK=off go build -o evosim

echo ""
echo "Running simulation with statistical analysis for 30 seconds..."
echo "Use [v] to cycle through views, especially 'statistical' and 'anomalies' views"
echo "Use [E] to export statistical data while running"
echo "Press [Q] to quit the simulation"
echo ""

# Run the simulation
timeout 30s ./evosim --pop-size 50 --seed 123 || true

echo ""
echo "Demo complete!"
echo ""
echo "The statistical analysis system has tracked:"
echo "• Every entity and plant energy change"
echo "• All births, deaths, and trait modifications"
echo "• System state snapshots every 10 ticks"
echo "• Anomaly detection every 50 ticks"
echo ""
echo "Key features demonstrated:"
echo "• Real-time anomaly detection (energy conservation, trait distributions)"
echo "• Statistical views in CLI (use [v] to cycle to 'statistical' and 'anomalies')"
echo "• Data export functionality ([E] key during simulation)"
echo "• Conservation law validation and biological plausibility checks"
echo ""
echo "Check for exported files: evosim_stats_*.csv and evosim_analysis_*.json"
ls -la evosim_* 2>/dev/null | head -5 || echo "No export files found (press [E] during simulation to export)"