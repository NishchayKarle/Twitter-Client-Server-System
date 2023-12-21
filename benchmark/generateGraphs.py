import matplotlib.pyplot as plt
import numpy as np


def divide(arr):
    val = arr[0]
    for i in range(len(arr)):
        arr[i] /= val
        arr[i] = 1 / arr[i]
        arr[i] = round(arr[i], 2)


def generateGraphs():
    f = open("time.txt", "r")
    xsmall = []
    small = []
    medium = []
    large = []
    xlarge = []

    threads = {1, 2, 4, 6, 8, 10, 12}

    for j in threads:
        time = 0
        for i in range(5):
            time += float(f.readline().strip("\n"))
        xsmall.append(time)

    for j in threads:
        time = 0
        for i in range(5):
            time += float(f.readline().strip("\n"))
        small.append(time)

    for j in threads:
        time = 0
        for i in range(5):
            time += float(f.readline().strip("\n"))
        medium.append(time)

    for j in threads:
        time = 0
        for i in range(5):
            time += float(f.readline().strip("\n"))
        large.append(time)

    for j in threads:
        time = 0
        for i in range(5):
            time += float(f.readline().strip("\n"))
        xlarge.append(time)

    f.close()

    divide(xsmall)
    divide(small)
    divide(medium)
    divide(large)
    divide(xlarge)

    xpoints = [1, 2, 4, 6, 8, 10, 12]
    plt.title("SpeedUp Graph")

    ypoints = np.array(xsmall)
    plt.plot(xpoints, ypoints, marker="o", label="XSMALL")

    ypoints = np.array(small)
    plt.plot(xpoints, ypoints, marker="o", label="SMALL")

    ypoints = np.array(medium)
    plt.plot(xpoints, ypoints, marker="o", label="MEDIUM")

    ypoints = np.array(large)
    plt.plot(xpoints, ypoints, marker="o", label="LARGE")

    ypoints = np.array(xlarge)
    plt.plot(xpoints, ypoints, marker="o", label="XLARGE")

    plt.legend()

    plt.xlabel("NUM OF THREADS")
    plt.ylabel("SPEEDUP")
    plt.savefig("SPEEDUP GRAPH.png")


if __name__ == "__main__":
    generateGraphs()
