#include <iostream>
#include <vector>
#include <stack>
#include <set>
#include <string>

using namespace std;

struct Position {
    int r, c;
    bool operator==(const Position& other) const { return r == other.r && c == other.c; }
    bool operator<(const Position& other) const { 
        if (r != other.r) return r < other.r; 
        return c < other.c; 
    }
};

class STCGrid {
public:
    int rows, cols;
    int megaRows, megaCols;
    vector<vector<bool>> blocked;
    vector<vector<vector<Position>>> treeEdges;

    STCGrid(int r, int c) {
        rows = (r % 2 == 0) ? r : r + 1;
        cols = (c % 2 == 0) ? c : c + 1;
        megaRows = rows / 2;
        megaCols = cols / 2;
        
        blocked.assign(rows, vector<bool>(cols, false));
        treeEdges.assign(megaRows, vector<vector<Position>>(megaCols));
    }

    void setObstacle(int mr, int mc) {
        if (mr >= 0 && mr < megaRows && mc >= 0 && mc < megaCols) {
            blocked[mr * 2][mc * 2] = true;
            blocked[mr * 2 + 1][mc * 2] = true;
            blocked[mr * 2][mc * 2 + 1] = true;
            blocked[mr * 2 + 1][mc * 2 + 1] = true;
        }
    }

    bool isMegaFree(int mr, int mc) {
        if (mr < 0 || mr >= megaRows || mc < 0 || mc >= megaCols) return false;
        return !blocked[mr * 2][mc * 2];
    }

    void buildSpanningTree(Position startMega) {
        set<Position> visited;
        stack<Position> s;
        
        s.push(startMega);
        visited.insert(startMega);

        int dr[] = {-1, 0, 1, 0}; // U, R, D, L
        int dc[] = {0, 1, 0, -1};

        while (!s.empty()) {
            Position curr = s.top();
            bool foundUnvisited = false;

            for (int i = 0; i < 4; i++) {
                Position next = {curr.r + dr[i], curr.c + dc[i]};
                
                if (isMegaFree(next.r, next.c) && visited.find(next) == visited.end()) {
                    treeEdges[curr.r][curr.c].push_back(next);
                    treeEdges[next.r][next.c].push_back(curr); 
                    
                    visited.insert(next);
                    s.push(next);
                    foundUnvisited = true;
                    break; 
                }
            }
            if (!foundUnvisited) s.pop();
        }
    }

    bool isValidMove(int r1, int c1, int r2, int c2) {
        if (r2 < 0 || r2 >= rows || c2 < 0 || c2 >= cols) return false;
        if (blocked[r2][c2]) return false;

        int mr1 = r1 / 2, mc1 = c1 / 2;
        int mr2 = r2 / 2, mc2 = c2 / 2;

        if (mr1 == mr2 && mc1 == mc2) return true;

        for (Position& neighbor : treeEdges[mr1][mc1]) {
            if (neighbor.r == mr2 && neighbor.c == mc2) return true;
        }
        return false;
    }

    void generatePath(Position startFine, Position targetFine) {
        cout << "[Start]  Fine-Cell (" << startFine.r << "," << startFine.c << ")\n";
        cout << "[Target] Fine-Cell (" << targetFine.r << "," << targetFine.c << ")\n";
        
        if (blocked[targetFine.r][targetFine.c]) {
            cout << "Target is inside an obstacle!\n";
            return;
        }

        int r = startFine.r;
        int c = startFine.c;
        int heading = 1; 
        int dirs[4][2] = {{-1, 0}, {0, 1}, {1, 0}, {0, -1}};
        char dirChars[] = {'U', 'R', 'D', 'L'};
        int turnOffsets[] = {1, 0, 3, 2}; 
        
        vector<char> pathDirections;
        vector<Position> pathCoords; 
        pathCoords.push_back({r, c}); // Record start position

        int maxSteps = rows * cols * 2; 
        int steps = 0;

        while (!(r == targetFine.r && c == targetFine.c) && steps < maxSteps) {
            bool moved = false;
            
            for (int offset : turnOffsets) {
                int tryHeading = (heading + offset) % 4;
                int nr = r + dirs[tryHeading][0];
                int nc = c + dirs[tryHeading][1];

                if (isValidMove(r, c, nr, nc)) {
                    heading = tryHeading; 
                    r = nr;               
                    c = nc;
                    pathDirections.push_back(dirChars[heading]);
                    pathCoords.push_back({r, c}); // Record new position
                    moved = true;
                    break;
                }
            }

            if (!moved) {
                cout << "Error: Robot got stuck!\n";
                return;
            }
            steps++;
        }

        if (steps >= maxSteps) {
             cout << "Error: Max steps reached. Target unreachable.\n";
             return;
        }

        cout << "\n[Success] Path to Target Found:\n";
        for (int i = 0; i < pathDirections.size(); i++) {
            cout << pathDirections[i] << (i == pathDirections.size() - 1 ? "" : " -> ");
        }
        cout << "\nTotal moves: " << pathDirections.size() << "\n";

        printVisualPathMap(startFine, targetFine, pathCoords);
    }

    void printEmptyGrid() {
        cout << "\n--- Initial Map ---\n";
        for (int r = 0; r < rows; r++) {
            for (int c = 0; c < cols; c++) {
                cout << (blocked[r][c] ? "# " : ". ");
            }
            cout << "\n";
        }
        cout << "-------------------\n\n";
    }

    void printVisualPathMap(Position startFine, Position targetFine, const vector<Position>& pathCoords) {
        vector<vector<char>> displayMap(rows, vector<char>(cols, '.'));

        // 1. Draw obstacles
        for (int r = 0; r < rows; r++) {
            for (int c = 0; c < cols; c++) {
                if (blocked[r][c]) displayMap[r][c] = '#';
            }
        }

        // 2. Draw path
        for (const Position& p : pathCoords) {
            displayMap[p.r][p.c] = '*';
        }

        // 3. Draw Start and Target (overwriting the path asterisk at those exact spots)
        displayMap[startFine.r][startFine.c] = 'S';
        displayMap[targetFine.r][targetFine.c] = 'T';

        cout << "\n--- Visited Path Map ---\n";
        for (int r = 0; r < rows; r++) {
            for (int c = 0; c < cols; c++) {
                cout << displayMap[r][c] << " ";
            }
            cout << "\n";
        }
        cout << "------------------------\n\n";
    }
};

int main() {
    STCGrid grid(6, 6);
    
    // Set obstacle in the center mega-cell
    grid.setObstacle(1, 1);
    
    grid.printEmptyGrid();
    grid.buildSpanningTree({0, 0});

    Position robotStart = {1, 1}; 
    Position target = {5, 0}; 
    
    grid.generatePath(robotStart, target);

    return 0;

    // run the code with:
    // g++ -std=c++17 stc.cpp -o stc

    // then do:
    // ./stc

}