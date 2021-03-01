function OnEvent(e)
    local subclass = e:getHeader("Event-Subclass") or ""
    if string.find(subclass, "sofia::") ~= 1 then return end
    local ip = e:getHeader("network_ip") or e:getHeader("network-ip")
    if not ip then return end
    local ua = e:getHeader("user-agent") or ""
    local to_user = e:getHeader("to-user") or ""
    local from_user = e:getHeader("from-user") or ""
    local auth_result = e:getHeader("auth-result") or ""
    local registration_type = e:getHeader("registration-type") or ""
    
    local s = string.format("*** %s, ip = %s, ua = %s, to = %s, from = %s, result = %s, type = %s\n", subclass, ip, ua, to_user, from_user, auth_result, registration_type)
    freeswitch.consoleLog("NOTICE", s)
    
    if subclass == "sofia::wrong_call_state" or subclass == "sofia::register_failure" then
        local cmd = "fail2ban-client set freeswitch banip " .. ip
        freeswitch.consoleLog("ERR", cmd .. "\n")
        --ÏÈ×¢ÊÍµô£¬Èç¹û×¢²áÊ§°Ü¾Í½ûÖ¹
        --os.execute(cmd)
    end
end

--freeswitch.consoleLog("INFO", "fail2ban.lua, ===\n" .. event:serialize() .. "===\n")
OnEvent(event)
