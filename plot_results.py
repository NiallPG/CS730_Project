#!/usr/bin/env python3
"""
Generate Figure 1 (discovery time) and Figure 2 (utilization) from results.csv.

Usage: python plot_results.py [results.csv]
Requires: pandas, matplotlib  (pip install pandas matplotlib)
"""
import sys
import pandas as pd
import matplotlib.pyplot as plt

CSV_PATH = sys.argv[1] if len(sys.argv) > 1 else 'results.csv'
GRID_SIZE = 100   # focus on 100x100 for the poster
K_FOR_UTIL = 5   # utilization comparison at this agent count (matches proposal Figure 2)

df = pd.read_csv(CSV_PATH)
df = df[df['rows'] == GRID_SIZE]

# ----- Summary table -----
summary = df.groupby(['algorithm', 'num_agents'])[
    ['time_to_discovery', 'makespan', 'mean_utilization', 'min_utilization']
].mean().round(2)
print(summary)
print()

# ----- Figure 1: discovery time vs num_agents -----
fig, ax = plt.subplots(figsize=(8, 5))

single = df[df['algorithm'] == 'single_agent_stc']
if len(single) > 0:
    ax.axhline(single['time_to_discovery'].mean(), color='gray',
               linestyle='--', label='Single Agent STC', linewidth=2)

for alg, label, marker, color in [
    ('voronoi_stc', 'Voronoi + STC', 'o', 'tab:orange'),
    ('darp_stc',    'DARP + STC',    's', 'tab:blue'),
]:
    means = (df[df['algorithm'] == alg]
             .groupby('num_agents')['time_to_discovery']
             .mean().sort_index())
    ax.plot(means.index, means.values, marker=marker, label=label,
            linewidth=2, markersize=10, color=color)

ax.set_xlabel('Number of Agents', fontsize=12)
ax.set_ylabel('Mean Time to Discovery (steps)', fontsize=12)
ax.set_title(f'Discovery Time vs. Agent Count ({GRID_SIZE}×{GRID_SIZE})', fontsize=13)
ax.legend(fontsize=11)
ax.grid(True, alpha=0.3)
fig.tight_layout()
fig.savefig('figure1_discovery.png', dpi=150)
print('wrote figure1_discovery.png')

# ----- Figure 2: utilization bars at k=K_FOR_UTIL -----
fig, ax = plt.subplots(figsize=(8, 5))

labels = ['Single Agent\nSTC']
mean_u = [1.0]
min_u  = [1.0]
for alg, label in [
    ('voronoi_stc', f'Voronoi + STC\n(k={K_FOR_UTIL})'),
    ('darp_stc',    f'DARP + STC\n(k={K_FOR_UTIL})'),
]:
    sub = df[(df['algorithm'] == alg) & (df['num_agents'] == K_FOR_UTIL)]
    labels.append(label)
    mean_u.append(sub['mean_utilization'].mean())
    min_u.append(sub['min_utilization'].mean())

x = range(len(labels))
width = 0.35
ax.bar([i - width/2 for i in x], mean_u, width, label='Mean Utilization', color='tab:blue')
ax.bar([i + width/2 for i in x], min_u,  width, label='Min Utilization',  color='tab:cyan')

ax.set_xticks(list(x))
ax.set_xticklabels(labels, fontsize=11)
ax.set_ylabel('Agent Utilization', fontsize=12)
ax.set_title(f'Agent Utilization ({GRID_SIZE}×{GRID_SIZE})', fontsize=13)
ax.set_ylim(0, 1.15)
ax.legend(fontsize=11)
ax.grid(True, alpha=0.3, axis='y')
fig.tight_layout()
fig.savefig('figure2_utilization.png', dpi=150)
print('wrote figure2_utilization.png')

# ----- Figure 3: per-seed DARP/Voronoi makespan ratio at k=5 -----
v = df[(df['algorithm']=='voronoi_stc') & (df['num_agents']==5)].set_index(['seed','target_idx'])
d = df[(df['algorithm']=='darp_stc')    & (df['num_agents']==5)].set_index(['seed','target_idx'])
joined = v.join(d, lsuffix='_v', rsuffix='_d').drop_duplicates(subset=['makespan_v','makespan_d'])
ratio = joined['makespan_d'] / joined['makespan_v']

fig, ax = plt.subplots(figsize=(8, 5))
ax.hist(ratio, bins=30, color='tab:blue', edgecolor='black', alpha=0.8)
ax.axvline(1.0, color='gray', linestyle='--', label='DARP = Voronoi')
ax.axvline(ratio.mean(), color='tab:red', linestyle='-', label=f'mean = {ratio.mean():.3f}')
ax.set_xlabel('DARP makespan / Voronoi makespan (per seed, k=5)', fontsize=12)
ax.set_ylabel('Number of seeds', fontsize=12)
ax.set_title(f'DARP vs Voronoi makespan ratio ({GRID_SIZE}×{GRID_SIZE}, k=5)', fontsize=13)
ax.legend(fontsize=11)
ax.grid(True, alpha=0.3, axis='y')
fig.tight_layout()
fig.savefig('figure3_ratio.png', dpi=150)
print('wrote figure3_ratio.png')