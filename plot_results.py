#!/usr/bin/env python3
"""
Generate paper figures from results_full.csv.

Produces:
  figure1_discovery.png   — discovery time vs agent count, one panel per grid size
  figure2_utilization.png — utilization bars, one panel per grid size
  figure3_partition_overhead.png — DARP vs Voronoi wall-clock partition time (log scale)
  figure4_makespan_ratio.png — DARP/Voronoi makespan ratio histogram

Usage: python plot_results.py [results_full.csv]
"""
import sys
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

CSV_PATH = sys.argv[1] if len(sys.argv) > 1 else "results_full.csv"
df = pd.read_csv(CSV_PATH)

SIZES = sorted(df["rows"].unique())
AGENTS = sorted(df[df["algorithm"] != "single_agent_stc"]["num_agents"].unique())

# ============================================================
# Summary table
# ============================================================
summary = (
    df.groupby(["rows", "algorithm", "num_agents"])[
        ["time_to_discovery", "makespan", "mean_utilization", "min_utilization"]
    ]
    .mean()
    .round(2)
)
print(summary)
print()

# ============================================================
# Figure 1: Discovery time vs agent count (one panel per size)
# ============================================================
fig, axes = plt.subplots(1, len(SIZES), figsize=(5.5 * len(SIZES), 5), sharey=False)
if len(SIZES) == 1:
    axes = [axes]

for ax, sz in zip(axes, SIZES):
    sub = df[df["rows"] == sz]

    sa = sub[sub["algorithm"] == "single_agent_stc"]
    if len(sa) > 0:
        ax.axhline(sa["time_to_discovery"].mean(), color="gray",
                   linestyle="--", label="Single Agent STC", linewidth=2)

    for alg, label, marker, color in [
        ("voronoi_stc", "Voronoi + STC", "o", "tab:orange"),
        ("darp_stc", "DARP + STC", "s", "tab:blue"),
    ]:
        means = (
            sub[sub["algorithm"] == alg]
            .groupby("num_agents")["time_to_discovery"]
            .mean()
            .sort_index()
        )
        ax.plot(means.index, means.values, marker=marker, label=label,
                linewidth=2, markersize=8, color=color)

    n_free = sa["makespan"].mean() if len(sa) > 0 else sub["makespan"].max()
    ideal_k = np.array(AGENTS)
    ax.plot(ideal_k, n_free / (2 * ideal_k), "g--", alpha=0.5, label="Ideal n/(2k)")

    ax.set_xlabel("Number of Agents", fontsize=12)
    ax.set_ylabel("Mean Time to Discovery (steps)", fontsize=12)
    ax.set_title(f"{sz}\u00d7{sz} grid", fontsize=13)
    ax.set_xticks(AGENTS)
    ax.legend(fontsize=9)
    ax.grid(True, alpha=0.3)

fig.suptitle("Discovery Time vs. Agent Count", fontsize=14, y=1.02)
fig.tight_layout()
fig.savefig("figure1_discovery.png", dpi=150, bbox_inches="tight")
print("wrote figure1_discovery.png")

# ============================================================
# Figure 2: Agent utilization (one panel per size)
# ============================================================
fig, axes = plt.subplots(1, len(SIZES), figsize=(5.5 * len(SIZES), 5), sharey=True)
if len(SIZES) == 1:
    axes = [axes]

K_FOR_UTIL = AGENTS[0]

for ax, sz in zip(axes, SIZES):
    sub = df[df["rows"] == sz]

    labels = ["Single\nAgent"]
    mean_u = [1.0]
    min_u = [1.0]

    for alg, lbl in [
        ("voronoi_stc", f"Voronoi\n(k={K_FOR_UTIL})"),
        ("darp_stc", f"DARP\n(k={K_FOR_UTIL})"),
    ]:
        s2 = sub[(sub["algorithm"] == alg) & (sub["num_agents"] == K_FOR_UTIL)]
        labels.append(lbl)
        mean_u.append(s2["mean_utilization"].mean())
        min_u.append(s2["min_utilization"].mean())

    x = np.arange(len(labels))
    width = 0.35
    ax.bar(x - width / 2, mean_u, width, label="Mean Util", color="tab:blue")
    ax.bar(x + width / 2, min_u, width, label="Min Util", color="tab:cyan")

    ax.set_xticks(x)
    ax.set_xticklabels(labels, fontsize=10)
    ax.set_ylabel("Utilization", fontsize=12)
    ax.set_title(f"{sz}\u00d7{sz} grid", fontsize=13)
    ax.set_ylim(0, 1.15)
    ax.legend(fontsize=9)
    ax.grid(True, alpha=0.3, axis="y")

fig.suptitle(f"Agent Utilization (k={K_FOR_UTIL})", fontsize=14, y=1.02)
fig.tight_layout()
fig.savefig("figure2_utilization.png", dpi=150, bbox_inches="tight")
print("wrote figure2_utilization.png")

# ============================================================
# Figure 3: Partition overhead — log scale, median + IQR
# ============================================================
multi = df[df["algorithm"].isin(["voronoi_stc", "darp_stc"])].copy()
timing = (
    multi.groupby(["algorithm", "rows", "num_agents", "seed"])["partition_time_ns"]
    .first()
    .reset_index()
)
timing["partition_time_ms"] = timing["partition_time_ns"] / 1e6

fig, ax = plt.subplots(figsize=(8, 5))

for alg, label, color, marker in [
    ("voronoi_stc", "Voronoi", "tab:orange", "o"),
    ("darp_stc", "DARP", "tab:blue", "s"),
]:
    t = timing[timing["algorithm"] == alg]
    grouped = t.groupby("rows")["partition_time_ms"]
    medians = grouped.median().sort_index()
    q25 = grouped.quantile(0.25).sort_index()
    q75 = grouped.quantile(0.75).sort_index()

    yerr_lo = (medians - q25).values
    yerr_hi = (q75 - medians).values

    ax.errorbar(medians.index, medians.values,
                yerr=[yerr_lo, yerr_hi],
                marker=marker, label=label, linewidth=2, markersize=8,
                color=color, capsize=4)

ax.set_yscale("log")
ax.set_xlabel("Grid Side Length", fontsize=12)
ax.set_ylabel("Partition Time (ms, log scale)", fontsize=12)
ax.set_title("Partitioning Overhead: DARP vs. Voronoi", fontsize=13)
ax.set_xticks(SIZES)
ax.set_xticklabels([f"{s}\u00d7{s}" for s in SIZES])
ax.legend(fontsize=11)
ax.grid(True, alpha=0.3, which="both")
fig.tight_layout()
fig.savefig("figure3_partition_overhead.png", dpi=150, bbox_inches="tight")
print("wrote figure3_partition_overhead.png")

print("\nPartition timing summary (ms):")
print(timing.groupby(["algorithm", "rows"])["partition_time_ms"].describe().round(3))

# ============================================================
# Figure 4: DARP/Voronoi makespan ratio histogram (per size)
# ============================================================
vor = (
    df[df["algorithm"] == "voronoi_stc"]
    .groupby(["rows", "num_agents", "seed"])["makespan"]
    .first()
    .reset_index()
    .rename(columns={"makespan": "vor_makespan"})
)
dar = (
    df[df["algorithm"] == "darp_stc"]
    .groupby(["rows", "num_agents", "seed"])["makespan"]
    .first()
    .reset_index()
    .rename(columns={"makespan": "darp_makespan"})
)
merged = vor.merge(dar, on=["rows", "num_agents", "seed"])
merged["ratio"] = merged["darp_makespan"] / merged["vor_makespan"]

fig, axes = plt.subplots(1, len(SIZES), figsize=(5.5 * len(SIZES), 5), sharey=False)
if len(SIZES) == 1:
    axes = [axes]

for ax, sz in zip(axes, SIZES):
    sub = merged[merged["rows"] == sz]
    ax.hist(sub["ratio"], bins=50, edgecolor="black", alpha=0.7, color="tab:blue")
    ax.axvline(1.0, color="red", linestyle="--", linewidth=1.5, label="ratio = 1.0")
    n_total = len(sub)
    n_tied = (sub["ratio"] == 1.0).sum()
    n_darp_wins = (sub["ratio"] < 1.0).sum()
    n_darp_loses = (sub["ratio"] > 1.0).sum()
    ax.set_xlabel("DARP Makespan / Voronoi Makespan", fontsize=11)
    ax.set_ylabel("Count (seed \u00d7 k)", fontsize=11)
    ax.set_title(
        f"{sz}\u00d7{sz}  \u2014  tied: {n_tied}/{n_total}, "
        f"wins: {n_darp_wins}, loses: {n_darp_loses}",
        fontsize=11,
    )
    ax.legend(fontsize=9)
    ax.grid(True, alpha=0.3, axis="y")

fig.suptitle("DARP vs. Voronoi Makespan Ratio", fontsize=14, y=1.02)
fig.tight_layout()
fig.savefig("figure4_makespan_ratio.png", dpi=150, bbox_inches="tight")
print("wrote figure4_makespan_ratio.png")

print("\nDone \u2014 4 figures written.")