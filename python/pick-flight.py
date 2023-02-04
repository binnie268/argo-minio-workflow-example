import pandas as pd

df2=pd.read_csv('/avged_flights.csv')
flight_dict = {"Flight": [], "Averaged Price": []}
df2["price_diff"] = df2["Averaged Price"].diff()
price_diff = 0
location = ""
for key, value in df2.iterrows():
    price_diff = value["price_diff"]
    if (price_diff > 0):
        location = df2.iloc[0]["Flight"]
    else:
        location = df2.iloc[1]["Flight"]

print("Pick {}!".format(location))
flight_dict = {"Decision": []}
flight_dict["Decision"].append(location)
# with open('flight_decision.txt', 'w') as f:
#     f.write("Pick {}!".format(location))
output_df = pd.DataFrame.from_dict(flight_dict)
output_df.to_csv(r'flight_decision.csv', index=False, header=True)