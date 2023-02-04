import pandas as pd
import os

print([x[0] for x in os.walk('./')])
print(os.path.isfile('flight_prices.csv'))
print(os.path.isfile('/flight_prices.csv'))
print(os.path.isfile('./flight_prices.csv'))
print(os.path.isfile('my-artifact.csv'))
print(os.path.isfile('./my-artifact.csv'))
print(os.path.isfile('/my-artifact.csv'))

df2=pd.read_csv('/my-artifact.csv')
flight_dict = {"Flight": [], "Sum": [], "Count": []}
grouped_obj = df2.groupby(["Flights"])

for key, item in grouped_obj:
    flight_dict["Flight"].append(key)
    flight_dict["Sum"].append(item["Price"].sum())
    flight_dict["Count"].append(item["Price"].count())
    print(flight_dict)
   
output_df = pd.DataFrame.from_dict(flight_dict)
output_df.to_csv(r'summed_flights.csv', index=False, header=True)