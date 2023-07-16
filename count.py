import sys
from collections import Counter

if len(sys.argv) < 3:
    print("Usage: python program.py <input file> <output file>")
    sys.exit(1)

filename = sys.argv[1]

with open(filename, "r") as file:
    lines = file.readlines()

kolom_pertama = [line.split(",")[0] for line in lines]
frekuensi = Counter(kolom_pertama)

total = 0
temp = 0
output_filename = sys.argv[2]

with open(output_filename, "w") as output_file:
    for freq in range(1, 502):
        total2 = 0
        if freq == 501:
            for lba, jumlah in frekuensi.items():
                if jumlah >= freq:
                    total2 += jumlah
            temp+=total2
            output_file.write(f">={freq} : {total2}\n")

        elif freq % 25 == 0:
            total2 = 0
            flag = freq - 24
            for lba, jumlah in frekuensi.items():
                if jumlah >= flag and jumlah <= freq:
                    total2 += jumlah
            temp += total2
            output_file.write(f"{flag}-{freq} : {total2}\n")

    output_file.write(f"Total keseluruhan: {temp}\n")

print("Output telah ditulis ke file:", output_filename)
