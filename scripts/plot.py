import csv
import os
import sys
import matplotlib.pyplot as plt
from matplotlib.patches import Patch

def main():
    csv_path = os.getenv('HISTOGRAM_CSV_PATH', 'results/histogram.csv')
    out_path = os.getenv('HISTOGRAM_IMG_PATH', 'results/histogram.png')

    if not os.path.exists(csv_path):
        print(f'Файл {csv_path} не найден. Пожалуйста, убедитесь, что он существует.')
        sys.exit(1)

    bytes_range = []
    cover = []
    stego = []

    with open(csv_path, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            bytes_range.append(int(row['byte']))
            cover.append(int(row['cover']))
            stego.append(int(row['stego']))

    diff = [stego[i] - cover[i] for i in range(len(cover))]

    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 6))
    configs = [
        (cover, 'Оригинал', 'blue', 0.5),
        (stego, 'Стегограмма', 'red', 0.5),
    ]

    for data, title, color, alpha in configs:
        ax1.bar(bytes_range, data, width=1.0, label=title, color=color, alpha=alpha, edgecolor='none')
        
    ax1.set_title("Оригинал против Стегограммы")
    ax1.set_xlabel('Байт')
    ax1.set_ylabel('Количество')
    ax1.set_xlim(0, 255)
    ax1.grid(True, linestyle='--', alpha=0.5)
    ax1.set_xticks(range(0,256,10))
    ax1.legend(
        loc='upper right',
        fontsize=11,
        frameon=True,
        framealpha=0.9,
        facecolor='white',
        edgecolor='gray'
    )

    bar_colors = ['red' if d >= 0 else 'blue' for d in diff]
    legend_elements = [
        Patch(facecolor='blue', label='Убыль Частоты', alpha=alpha),
        Patch(facecolor='red', label='Прирост частоты', alpha=alpha)
    ]
    ax2.bar(bytes_range, diff, width=1.0, label=title, color=bar_colors, alpha=alpha, edgecolor='none')
    ax2.set_title("Изменение частот")
    ax2.set_xlabel('Байт')
    ax2.set_ylabel('изменение')
    ax2.set_xlim(0, 255)
    ax2.grid(True, linestyle='--', alpha=0.5)
    ax2.axhline(0, color='black', linewidth=0.8, linestyle='-')
    ax2.set_xticks(range(0,256,10))
    ax2.legend(
        handles=legend_elements,
        loc='upper right',
        fontsize=11,
        frameon=True,
        framealpha=0.9,
        facecolor='white',
        edgecolor='gray'
    )

    plt.tight_layout()
    plt.savefig(out_path, dpi=300)
    print(f'\nГистограмма сохранена в {out_path}')

if __name__ == '__main__':
    main()