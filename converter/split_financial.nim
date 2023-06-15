import parsecsv
import strutils
from os import paramStr
from streams import newFileStream, writeLine, close
import tables

const BSIZE=512

var cbz: CountTable[int]
var s = newFileStream(paramStr(1), fmRead)
if s==nil:
  quit("cannot open the file" & paramStr(1))

var wr = newFileStream(paramStr(2), fmWrite)
if wr==nil:
  quit("cannot write to file" & paramStr(2))

var x : CsvParser
open(x, s, paramStr(1))

var 
  asu : string
  lba : int
  size : int
  opcode : string
  nondivide : int = 0
  tambah : int

var res : bool
res=readRow(x)
while res:
  var num = 0
  for val in items(x.row):
    case num
    of 0:
      discard
    of 1:
      lba = parseInt(val)
      if lba<0:
        break
    of 2:
      size = parseInt(val)
      tambah = 0
      if size mod BSIZE > 0:
        tambah = 1
      size = size div BSIZE
    of 3:
      opcode = val
      for i in countdown(tambah+size,1):
        wr.writeLine(lba, ",", toUpper(opcode))
        inc(lba, 1)
    else:
      break
    inc(num)

  res=readRow(x)
close(x)
close(s)
close(wr)
