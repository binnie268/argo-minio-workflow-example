import pandas as pd
import os

print([x[0] for x in os.walk('./')])
print(os.path.isfile('summed_flights.csv'))
print(os.path.isfile('./summed_flights.csv'))
print(os.path.isfile('/summed_flights.csv'))


df2=pd.read_csv('/summed_flights.csv')
flight_dict = {"Flight": [], "Averaged Price": []}

for key, value in df2.iterrows():
    flight_dict["Flight"].append(value["Flight"])
    flight_dict["Averaged Price"].append(value["Sum"]//value["Count"])
    print(flight_dict)
   
output_df = pd.DataFrame.from_dict(flight_dict)
output_df.to_csv(r'avged_flights.csv', index=False, header=True)