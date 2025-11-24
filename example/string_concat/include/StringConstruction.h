#ifndef STRINGCONSTRUCTION_H
#define STRINGCONSTRUCTION_H

#include <string>
using std::string;




#include <unordered_map>

const int INF = 2147483647;
int minCloneLength;

int min(int a, int b) {
	if (a < b) {
		return a;
	}
	else {
		return b;
	}
}

int findMaxRepeatedLength(size_t index, const std::string& target) {
    size_t n = target.length();
    
    size_t maxPossibleLength = std::min(index, n - index);
    size_t maxLen = 0;

    for (size_t left = 0; left < index; ++left) {
		if ((index - left) <= maxLen) {
            break;
        }

        if (target[left] == target[index]) {
            size_t currentLen = 1;

            while (currentLen < maxPossibleLength && 
                   left + currentLen < index && 
                   target[left + currentLen] == target[index + currentLen]) {
                currentLen++;
            }

            if (currentLen > maxLen) {
                maxLen = currentLen;
    
                if (maxLen == maxPossibleLength) return maxLen;
            }
        }
    }

    return (maxLen >= minCloneLength) ? maxLen : 0;
}

int findCost(size_t index, int cloneCost, int appendCost, std::string& target, std::unordered_map<int, int>& minMap) {
	if (target.size() == index) {
		return 0;
	}

	if (minMap.find(index) != minMap.end()) {
		return minMap[index];
	}

	int cloneVal = INF;
	int appendVal;
	int minVal;

	int len = findMaxRepeatedLength(index, target);

	if (len != 0) {
		cloneVal = cloneCost + findCost(index + len, cloneCost, appendCost, target, minMap);	
	}

	appendVal = appendCost + findCost(index + 1, cloneCost, appendCost, target, minMap);


	minVal = min(cloneVal, appendVal);

	minMap[index] = minVal;

	return minVal;
}

int stringConstruction(std::string target, int cloneCost, int appendCost) {
	std::unordered_map<int, int> minMap;
	minCloneLength = (cloneCost + appendCost - 1) / appendCost;

	return findCost(0, cloneCost, appendCost, target, minMap);
}



#endif