# -*- mode: ruby -*-
# vi: set ft=ruby :

#
# Ensure you have followed the instructions to allow 
# memory overcommit before running "vagrant up".
#
# http://www.virtuallyghetto.com/2015/02/quick-tip-how-to-enable-memory-overcommitment-in-vmware-fusion.html
#

Vagrant.configure("2") do |config|

    config.vm.define "nfs" do |nfs|
        nfs.vm.box = "mevansam/ubuntu1404-nfsserver"
        nfs.vm.hostname = "nfsbox"
        nfs.vm.synced_folder ".", "/vagrant", disabled: true
        
        nfs.ssh.insert_key = false
        nfs.ssh.username = "ubuntu"
        nfs.ssh.password = "Vmware6!"
        
        nfs.vm.provider :vmware_fusion do |vm|
            
            vm.vmx["memsize"] = "1024"
            vm.vmx["numvcpus"] = "1"
        end
    end
    
    config.vm.define "esx" do |esx|

        esx.vm.box = "mevansam/vmware-esxi6-vc"
        esx.vm.hostname = "esxbox"
        esx.vm.synced_folder ".", "/vagrant", disabled: true
        esx.vm.network "forwarded_port", guest: 443, host: 18443
        
        esx.vm.network "public_network", ip: "172.16.139.132"
        
        esx.ssh.insert_key = false
        esx.ssh.username = "root"
        esx.ssh.password = "Vmware6!"
        esx.ssh.shell = "sh"
        config.ssh.port = 22
        
        esx.vm.provider :vmware_fusion do |vm|
            
            vm.vmx["memsize"] = "6144"
            vm.vmx["numvcpus"] = "2"
            vm.vmx["vhv.enable"] = "TRUE"            
            vm.vmx["hard-disk.hostBuffer"] = "disabled"
        end

    end
end
