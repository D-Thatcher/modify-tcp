sudo iptables -L --line-numbers
sudo iptables -D INPUT 1
sudo iptables -I INPUT -d 192.168.0.0/24 -j NFQUEUE --queue-num 1
sudo ufw disable
sudo iptables -A INPUT -i wlp0s20f3 -j NFQUEUE --queue-num 1
sudo ufw enable
sudo iptables -D INPUT -i wlp0s20f3 -j NFQUEUE --queue-num 1

sudo /home/ubuntu18/miniconda3/bin/python3 main.py

iptables -t nat -A PREROUTING -p tcp --dport 80 -j NFQUEUE --queue-num 1