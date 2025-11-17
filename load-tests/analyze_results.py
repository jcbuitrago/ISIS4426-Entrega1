#!/usr/bin/env python3
"""
Analiza el archivo results.json generado por k6 y muestra un resumen legible.
Uso: python analyze_results.py results.json
"""
import json
import sys
from collections import defaultdict
from datetime import datetime

def analyze_k6_results(json_file):
    print(f"\n{'='*70}")
    print(f"  Análisis de resultados k6: {json_file}")
    print(f"{'='*70}\n")
    
    metrics = defaultdict(list)
    checks = defaultdict(lambda: {'passed': 0, 'failed': 0})
    http_reqs = []
    scenarios_data = defaultdict(list)
    
    with open(json_file, 'r', encoding='utf-8') as f:
        for line in f:
            try:
                data = json.loads(line.strip())
                
                # Métricas finales
                if data.get('type') == 'Point':
                    metric_name = data.get('metric')
                    value = data.get('data', {}).get('value', 0)
                    tags = data.get('data', {}).get('tags', {})
                    
                    if metric_name:
                        metrics[metric_name].append(value)
                        
                    # Agrupar por escenario
                    scenario = tags.get('scenario')
                    if scenario and metric_name == 'http_req_duration':
                        scenarios_data[scenario].append(value)
                    
                    # Requests HTTP
                    if metric_name == 'http_reqs':
                        http_reqs.append(data)
                        
                    # Checks
                    if metric_name == 'checks':
                        check_name = tags.get('check', 'unknown')
                        if value == 1:
                            checks[check_name]['passed'] += 1
                        else:
                            checks[check_name]['failed'] += 1
                            
            except json.JSONDecodeError:
                continue
    
    # Resumen de HTTP requests
    if 'http_reqs' in metrics:
        total_reqs = len(metrics['http_reqs'])
        print(f"📊 REQUESTS TOTALES: {total_reqs:,}")
    
    # Duración HTTP
    if 'http_req_duration' in metrics:
        durations = sorted(metrics['http_req_duration'])
        if durations:
            avg = sum(durations) / len(durations)
            p50 = durations[int(len(durations) * 0.50)]
            p95 = durations[int(len(durations) * 0.95)]
            p99 = durations[int(len(durations) * 0.99)]
            max_dur = max(durations)
            
            print(f"\n⏱️  LATENCIAS (ms):")
            print(f"   Promedio: {avg:.2f} ms")
            print(f"   p(50):    {p50:.2f} ms")
            print(f"   p(95):    {p95:.2f} ms  {'✅' if p95 < 800 else '⚠️'}")
            print(f"   p(99):    {p99:.2f} ms")
            print(f"   Max:      {max_dur:.2f} ms")
    
    # Tasa de fallos
    if 'http_req_failed' in metrics:
        failed = sum(metrics['http_req_failed'])
        total = len(metrics['http_req_failed'])
        fail_rate = (failed / total * 100) if total > 0 else 0
        status = '✅' if fail_rate < 2 else '❌'
        print(f"\n🚨 TASA DE FALLOS: {fail_rate:.2f}% {status}")
        print(f"   Fallidos: {int(failed):,} de {total:,}")
    
    # Métricas personalizadas
    custom_metrics = ['upload_duration', 's3_asset_duration', 'vote_failures']
    for metric in custom_metrics:
        if metric in metrics:
            values = [v for v in metrics[metric] if v > 0]
            if values:
                print(f"\n📈 {metric.upper().replace('_', ' ')}:")
                avg = sum(values) / len(values)
                p95 = sorted(values)[int(len(values) * 0.95)]
                print(f"   Promedio: {avg:.2f}")
                print(f"   p(95):    {p95:.2f}")
                if metric == 'vote_failures':
                    print(f"   Total:    {int(sum(values))}")
    
    # Resultados por escenario
    if scenarios_data:
        print(f"\n🎯 LATENCIAS POR ESCENARIO:")
        for scenario, durations in scenarios_data.items():
            if durations:
                sorted_durs = sorted(durations)
                avg = sum(sorted_durs) / len(sorted_durs)
                p95 = sorted_durs[int(len(sorted_durs) * 0.95)]
                print(f"\n   {scenario}:")
                print(f"      Requests: {len(durations):,}")
                print(f"      Avg:      {avg:.2f} ms")
                print(f"      p(95):    {p95:.2f} ms")
    
    # Checks
    if checks:
        print(f"\n✓ CHECKS:")
        total_passed = sum(c['passed'] for c in checks.values())
        total_failed = sum(c['failed'] for c in checks.values())
        total_checks = total_passed + total_failed
        success_rate = (total_passed / total_checks * 100) if total_checks > 0 else 0
        
        print(f"   Total: {total_checks:,} ({success_rate:.1f}% exitosos)")
        print(f"   ✅ Pasados: {total_passed:,}")
        print(f"   ❌ Fallidos: {total_failed:,}")
        
        # Detalles de checks fallidos
        if total_failed > 0:
            print(f"\n   Checks con fallos:")
            for check_name, counts in checks.items():
                if counts['failed'] > 0:
                    fail_pct = (counts['failed'] / (counts['passed'] + counts['failed']) * 100)
                    print(f"      • {check_name}: {counts['failed']} fallos ({fail_pct:.1f}%)")
    
    # VUs (usuarios virtuales)
    if 'vus' in metrics:
        max_vus = max(metrics['vus']) if metrics['vus'] else 0
        print(f"\n👥 VUs MÁXIMOS: {int(max_vus)}")
    
    # Data received/sent
    if 'data_received' in metrics:
        total_received = sum(metrics['data_received'])
        total_mb = total_received / (1024 * 1024)
        print(f"\n📥 DATA RECIBIDA: {total_mb:.2f} MB")
    
    if 'data_sent' in metrics:
        total_sent = sum(metrics['data_sent'])
        total_mb = total_sent / (1024 * 1024)
        print(f"📤 DATA ENVIADA: {total_mb:.2f} MB")
    
    print(f"\n{'='*70}\n")

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Uso: python analyze_results.py <results.json>")
        sys.exit(1)
    
    json_file = sys.argv[1]
    try:
        analyze_k6_results(json_file)
    except FileNotFoundError:
        print(f"❌ Error: No se encontró el archivo {json_file}")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error al procesar: {e}")
        sys.exit(1)


