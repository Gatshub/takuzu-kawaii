# Takuzu kawaii ♡

Puzzle de logique binaire (a.k.a. Binairo) dans le terminal, en Go + Bubble Tea.

## Règles
- Remplir la grille de ○ (0) et ● (1) :
  - autant de 0 que de 1 sur chaque ligne et colonne ;
  - jamais trois chiffres identiques côte à côte ;
  - pas deux lignes (ni colonnes) identiques.

## Niveaux
| Niveau    | Grille | Indices |
|-----------|--------|---------|
| facile    | 6×6    | 2       |
| moyen     | 8×8    | 1       |
| difficile | 10×10  | 0       |

## Touches
`↑↓←→` bouger · `espace`/`0`/`1` placer · `backspace` effacer · `n` rejouer · `m` menu · `q` quitter

## Couleurs
Palette adaptative (lipgloss AdaptiveColor) : suit le thème clair ou sombre du terminal.
Gouttières ✓/✗ : état de chaque ligne et colonne en temps réel.

## Build
    go build -o takuzu .
