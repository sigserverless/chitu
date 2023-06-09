from Dserver.dserver import DServer

print("Hello", flush=True)

dserver = DServer()
ddict1 = dserver.new_D("ddict1")

ddict2 = dserver.new_D("ddict2")
ddict3 = dserver.merge_Ds(ddict1, ddict2)

ddict1.set("a", [42, 43, 44])
ddict2.set("a", [51, 52, 53])

ddict1.end()
ddict2.end()

ddict1_res = ddict1.wait()
print(ddict1_res, flush=True)

ddict3_res = ddict3.wait()

print(ddict3_res, flush=True)
dserver.close()